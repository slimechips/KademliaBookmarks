package node

import (
	"bufio"
	"container/list"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type NodeCore struct {
	GUID ID
	IP   net.IP
	Port int
}

func (node NodeCore) String() string {
	return fmt.Sprintf("%s~%s~%d~;", node.GUID.String(), node.IP.String(), node.Port)
}

type Node struct {
	NodeCore     NodeCore
	RoutingTable RoutingTable
	Data         map[ID]string //stores a <key,value> pair for retrieval
	Alive        bool          //for unit testing during prototyping
	mutex        *sync.Mutex
}

func (n *Node) Update(otherNodeCore *NodeCore) {

	prefix_length := otherNodeCore.GUID.Xor(n.NodeCore.GUID).PrefixLen()

	bucket := n.RoutingTable.Buckets[prefix_length]

	var element *list.Element = nil
	for e := bucket.Front(); e != nil; e = e.Next() {
		if e.Value.(*NodeCore).GUID.Equals((otherNodeCore).GUID) {
			element = e
			break
		}
	}
	n.mutex.Lock()
	if element == nil {
		if bucket.Len() <= BUCKET_SIZE {
			bucket.PushFront(otherNodeCore)
		} else {
			// TODO: Handle insertion when the list is full by evicting old elements if

			LRUNode := bucket.Back()
			ping := true
			if !ping {
				// remove the LRUNode
				bucket.Remove(LRUNode)

				// add the new node to the front
				bucket.PushFront(otherNodeCore)
			}
		}
	} else {
		bucket.MoveToFront(element)
	}
	n.mutex.Unlock()
}

func (node Node) FullAddr() string {
	return fmt.Sprintf("%s:%d", node.NodeCore.IP.String(), node.NodeCore.Port)
}

func (n *Node) FindClosest(target ID, count int) (ret []NodeCore) {
	records := n.FindClosestRecord(target, count)
	for i := range records {
		ret = append(ret, *records[i].Node)
	}
	return
}

func (n *Node) FindClosestRecord(target ID, count int) (ret []nodeCoreRecord) {
	bucket_num := target.Xor(n.NodeCore.GUID).PrefixLen()
	bucket := n.RoutingTable.Buckets[bucket_num]
	copyToList(bucket, &ret, target)

	// bidirectional search from the middle node
	for i := 1; (bucket_num-i >= 0 || bucket_num+i < ID_LENGTH*8) && len(ret) < count; i++ {
		if bucket_num-i >= 0 {
			bucket = n.RoutingTable.Buckets[bucket_num-i]
			copyToList(bucket, &ret, target)
		}
		if bucket_num+1 < ID_LENGTH*8 {
			bucket = n.RoutingTable.Buckets[bucket_num+i]
			copyToList(bucket, &ret, target)
		}
	}

	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].SortKey.Less(ret[j].SortKey)
	})

	// slice if somehow the list is longer than what is needed
	if len(ret) > count {
		ret = ret[:count]
	}

	return ret
}

//recvJoinReq appends new nodecore to my routing table and send listRecv to this new nodecore
func (node *Node) recvJoinReq(nodecore *NodeCore, conn *net.UDPConn, addr *net.UDPAddr) {
	prefix_length := nodecore.GUID.Xor(node.NodeCore.GUID).PrefixLen()

	node.SendResponse(LIST_MSG, convertKBucketToString(node.RoutingTable.Buckets[prefix_length]), conn, addr)
	node.Update(nodecore)
	log.Println("UPDATED", nodecore.GUID.String(), prefix_length)
}

func (node *Node) recvJoinListNodeCore(list []*NodeCore) {
	for _, nodeCore := range list {
		node.Update(nodeCore)
	}
}

func (node *Node) recvPing(conn *net.UDPConn, addr *net.UDPAddr) {
	go node.SendResponse(PING_MSG, "hi", conn, addr)
}

func (node *Node) recvStore(key ID, value string) {
	node.Data[key] = value
}

func (node *Node) recvTryFindKeyValue(key ID, conn *net.UDPConn, addr *net.UDPAddr) {
	if val, ok := node.Data[key]; ok {
		go node.SendResponse(FVALUE_MSG, val, conn, addr)
		return
	}
	nodeCores := node.FindClosest(key, K)
	s := ""
	for _, nodeCore := range nodeCores {
		log.Println("nodeCore found:", nodeCore.String())
		s += nodeCore.String()
	}
	go node.SendResponse(FVALUEFAIL_MSG, s, conn, addr)
}

func (node *Node) recvFindNode(key ID, conn *net.UDPConn, addr *net.UDPAddr) {
	nodeCores := node.FindClosest(key, K)
	s := ""
	for _, nodeCore := range nodeCores {
		log.Println("nodeCore found:", nodeCore.String())
		s += nodeCore.String()
	}
	log.Println("nodeCores all found:", s)
	go node.SendResponse(FLOOKUP_MSG, s, conn, addr)
}

//KNodesLookUp returns K number of NodeCore for storage of target key
func (node *Node) KNodesLookUp(key ID) []*NodeCore {
	chanFail := make(chan string)
	chanSucc := make(chan string)
	nodeCores := node.FindClosest(key, K)

	alive := make([]string, 0)
	requested := make([]string, 0)
	for i := 0; i < len(nodeCores); i++ {
		log.Println("TO REQUEST --> ", nodeCores[i].String())
		go node.Send(&nodeCores[i], FLOOKUP_MSG, key.String(), chanFail)
		requested = append(requested, nodeCores[i].String())
	}
	timer := time.NewTimer(time.Duration(TIMEOUT_DURATION))
iterativeFind:
	for {
		select {
		case <-timer.C:
			break iterativeFind
		case msg := <-chanFail:
			timer = time.NewTimer(time.Duration(TIMEOUT_DURATION))
			s := strings.Split(msg, "#")
			alive = append(alive, s[0])
			nodeCoreList := convertStringToNodeCoreList(strings.Split(s[1], ";"))
			for _, n := range nodeCoreList {
				if !StringsListContains(n.GUID.String(), requested) {
					requested = append(requested, n.String())
					go node.Send(n, FLOOKUP_MSG, key.String(), chanFail, chanSucc)
				}
			}
		case <-chanSucc:
			break iterativeFind
		}
	}
	alive = append(alive, node.NodeCore.String())
	resultNCList := convertStringToNodeCoreList(alive)
	sort.SliceStable(resultNCList, func(i, j int) bool {
		return resultNCList[i].GUID.Xor(key).Less(resultNCList[j].GUID.Xor(key))
	})
	if len(resultNCList) < K {
		return resultNCList
	} else {
		return resultNCList[:K]
	}
}

func (node *Node) StoreInNodes(nodeCores []*NodeCore, key ID, val string) {
	for _, nodeCore := range nodeCores {
		if nodeCore.GUID.Equals(node.NodeCore.GUID) {
			go node.recvStore(key, val)
		} else {
			go node.Send(nodeCore, STORE_MSG, fmt.Sprintf("%s;%s;", key.String(), val), nil)
		}
	}
}

//FindValueByKey sends request to Nodes for target key-value
func (node *Node) FindValueByKey(key ID) {
	if val, ok := node.Data[key]; ok {
		log.Println(val)
		return
	}
	chanSucc := make(chan string)
	chanFail := make(chan string)
	nodeCores := node.FindClosest(key, K)
	requested := make([]string, 0)
	for i := 0; i < len(nodeCores); i++ {
		go node.Send(&nodeCores[i], FVALUE_MSG, key.String(), chanFail, chanSucc)
		requested = append(requested, nodeCores[i].String())
	}
	timer := time.NewTimer(time.Duration(TIMEOUT_DURATION))
iterativeFind:
	for {
		select {
		case <-timer.C:
			break iterativeFind
		case msg := <-chanFail:
			timer = time.NewTimer(time.Duration(TIMEOUT_DURATION))
			s := strings.Split(msg, "#")
			nodeCoreList := convertStringToNodeCoreList(strings.Split(s[1], ";"))
			for _, n := range nodeCoreList {
				if !StringsListContains(n.GUID.String(), requested) {
					go node.Send(n, FVALUE_MSG, key.String(), chanFail, chanSucc)
					requested = append(requested, n.String())
				}
			}
		case <-chanSucc:
			break iterativeFind
		}
	}
}

/*
format: <senderNodeID>|<IP_address>|<tag>|<msgContent>|
Tags: join, list, ping, store, fValue, fNode
*/
func (node *Node) Send(nodeCore *NodeCore, tag string, rawMsg string, opts ...chan string) {
	// Dont send to yourself
	if node.NodeCore.IP.String() == nodeCore.IP.String() {
		return
	}
	addr := fmt.Sprintf("%s:%d", nodeCore.IP.String(), nodeCore.Port)
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("Some error %v", err)
		return
	}

	msg := fmt.Sprintf("%s|%s|%s|%s|", node.NodeCore.GUID.String(), node.NodeCore.IP.String(), tag, rawMsg)
	log.Printf("Node %d->Sending to %s ---->----- %s\n", node.NodeCore.GUID, addr, msg)
	fmt.Fprintf(conn, msg)
	//waiting for response
	_, err = bufio.NewReader(conn).Read(p)
	node.ResponseMsgHandler(fmt.Sprintf("~%s~%d~", nodeCore.IP.String(), nodeCore.Port), string(p), opts)
	if err == nil {
		log.Printf("%s\n", p)
	} else {
		log.Printf("Some error %v\n", err)
	}
	conn.Close()
}

/*
Send message to given address
Port suffix is added automatically
Message format: <NODE_ID>|<TAG>|<MSG>
Tags: Hello, AddPeer
*/
func (node *Node) AddrSend(addr string, tag string, rawMsg string) {
	// Dont send to yourself
	if node.NodeCore.IP.String() == addr {
		return
	}
	addr = fmt.Sprintf("%s:%d", addr, RECEIVER_PORT)
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("Some error %v", err)
		return
	}

	msg := fmt.Sprintf("%s|%s|%s|%s|", node.NodeCore.GUID.String(), node.NodeCore.IP.String(), tag, rawMsg)
	log.Printf("Node %s-Sending to %s msg --->-- %s\n", node.NodeCore.GUID.String(), addr, msg)
	fmt.Fprintf(conn, msg)
	_, err = bufio.NewReader(conn).Read(p)
	ad := strings.Split(addr, ":")
	node.ResponseMsgHandler(fmt.Sprintf("~%s~%s~", ad[0], ad[1]), string(p), nil)
	if err == nil {
		log.Printf("%s\n", p)
	} else {
		log.Printf("Some error %v\n", err)
	}
	conn.Close()
}
func (node *Node) SendResponse(tag string, rawMsg string, conn *net.UDPConn, addr *net.UDPAddr) {
	msg := fmt.Sprintf("%s|%s|%s|%s|", node.NodeCore.GUID.String(), node.NodeCore.IP.String(), tag, rawMsg)
	log.Println("SENDING RESPONSE:", msg)
	_, err := conn.WriteToUDP([]byte(msg), addr)
	if err != nil {
		log.Printf("Couldn't send response %v", err)
	}
}

func createNodeCore(guid ID, ip_address net.IP, udp_port int) *NodeCore {
	n := NodeCore{GUID: guid, IP: ip_address, Port: udp_port}
	return &n
}

//NewNode initalise node based on system IP
func NewNode(alive bool) *Node {
	ip := getIp()
	id := getNodeName(ip)
	node := &Node{
		NodeCore:     *createNodeCore(id, ip, RECEIVER_PORT),
		RoutingTable: *NewRoutingTable(),
		Data:         make(map[ID]string),
		Alive:        alive,
		mutex:        &sync.Mutex{},
	}
	log.Println("Current Node Info:", *node)
	return node
}

func (node *Node) StartListening() {
	addr := net.UDPAddr{
		Port: node.NodeCore.Port,
		IP:   node.NodeCore.IP,
	}
	log.Printf("%v-Started listening\n", node.FullAddr())
	// Listen on localhost
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Some error %v\n", err)
		return
	}
	for {
		//msgHandler:
		//format: <senderNodeID>|<IP_address>|<tag>|<msgContent>|
		//1. receive joinReq: //append new nodecore to my routing table and send listRecv to this new nodecore
		//2. receive listRecv://append routing table of node core to my routing table
		//3. receive pingRecv://send" yes im alive"
		//4. receive storeRecv://store keyvalue in data -> map
		//5. receive findValue://check my store for key:value and send kv to sender if not inside XOR(myid, target key) and send list in associated kbucket
		//6. receive findNode(target Key):XOR(my id, target key) send list of nodecore in assc kbucket
		// Byte array to hold message
		p := make([]byte, 2048)
		_, remoteAddr, err := ser.ReadFromUDP(p)
		log.Printf("Received msg: %s", string(p))
		if err != nil {

			log.Printf("Some error  %v\n", err)
			return
		}

		t := strings.Split(string(p), "|")
		if len(t) < 4 {
			log.Println("Error, received invalid message format")
			continue
		}
		senderID := ConvertStringToID(t[0])
		IP_address := net.ParseIP(t[1])
		tag := t[2]
		msgContent := t[3]
		log.Println("RECEIVING:", msgContent)
		// Receiver perspective of what request send it
		switch tag {
		case JOIN_MSG:
			go node.recvJoinReq(createNodeCore(senderID, IP_address, RECEIVER_PORT), ser, remoteAddr)
		case PING_MSG:
			go node.recvPing(ser, remoteAddr)
		case STORE_MSG:
			kv := strings.Split(msgContent, ";")
			go node.recvStore(ConvertStringToID(kv[0]), kv[1])
		case FVALUE_MSG:
			go node.recvTryFindKeyValue(ConvertStringToID(msgContent), ser, remoteAddr)
		case FLOOKUP_MSG:
			go node.recvFindNode(ConvertStringToID(msgContent), ser, remoteAddr)
		default: // e.g. hello
			// log.Printf("Node %d-Recv Msg from Node %d(%s): %s \n", node.ID.String(), senderID, senderIp.String(), recvMsg)
			// if node.tryAddtoNetwork(senderID, senderIp) {
			// 	node.broadcastNewPeer(senderID, senderIp)
			// }
		}
	}
}

func NodeCoreListContains(nc NodeCore, ncList *[]NodeCore) bool {
	for _, n := range *ncList {
		if n.GUID == nc.GUID {
			return true
		}
	}
	return false
}

func convertKBucketToString(bucket *list.List) string {
	s := ""
	for e := bucket.Front(); e != nil; e = e.Next() {
		NodeCore := e.Value.(*NodeCore)
		s += NodeCore.String()
	}
	return s
}

func ConvertStringToID(s string) ID {
	i, _ := hex.DecodeString(s)
	var id_string ID
	copy(id_string[:ID_LENGTH], i)
	return id_string
}

func StringsListContains(s string, stringList []string) bool {
	for _, str := range stringList {
		if s == str {
			return true
		}
	}
	return false
}

func convertStringToNodeCoreList(stringList []string) []*NodeCore {
	nodecoreList := make([]*NodeCore, 0)
	for i := 0; i < len(stringList)-1; i++ {
		//log.Println("CONVERTING TO NODECORE", stringList[i])
		s1 := strings.Split(stringList[i], "~")
		id := ConvertStringToID(s1[0])
		Port, _ := strconv.Atoi(s1[2])
		nodecoreList = append(nodecoreList,
			createNodeCore(id, net.ParseIP(s1[1]), Port))
	}
	return nodecoreList
}

func (node *Node) ResponseMsgHandler(recvAddr string, s string, chans []chan string) {
	t := strings.Split(s, "|")
	if len(t) < 4 {
		//TODO: handle error better
		log.Println("Error, received invalid message format")
		return
	}
	senderID := ConvertStringToID(t[0])
	//IP_address := net.ParseIP(t[1])
	tag := t[2]
	msgContent := t[3]
	log.Println("RESPONSE_HANDLER:", senderID, recvAddr, msgContent)
	switch tag {
	case PING_MSG:
		//TODO: pass (drop)
	case FVALUE_MSG:
		chans[1] <- senderID.String() + recvAddr + "#" + msgContent
		log.Println(msgContent)
	case FVALUEFAIL_MSG:
		chans[0] <- senderID.String() + recvAddr + "#" + msgContent
	case FLOOKUP_MSG:
		chans[0] <- senderID.String() + recvAddr + "#" + msgContent
	case LIST_MSG:
		if strings.Contains(msgContent, ";") {
			nodeList := strings.Split(msgContent, ";")
			go node.recvJoinListNodeCore(convertStringToNodeCoreList(nodeList))
		}
	}
}

// /*
// Broadcast a new peer to existing peers you know
// */
// func (node *Node) broadcastNewPeer(id int, ip net.IP) {
// 	for _, peer := range node.Peers {
// 		log.Printf("Node %d-Sending to %s - New Peer Id: %d New Peer Ip: %s\n",
// 			node.ID.String(), peer.FullAddr(), id, ip.String())

// 		msg := fmt.Sprintf("%d;%s;", id, ip.String())
// 		go node.Send(peer.IP.String(), "AddPeer", msg)
// 	}
// }

/*
Get IP address of current node
Works only in linux/docker
*/
func getIp() net.IP {
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		if i.Name != "eth0" {
			continue
		}
		addrs, _ := i.Addrs()
		var ip net.IP
		switch v := addrs[0].(type) {
		case *net.IPAddr:
			ip = v.IP
		case *net.IPNet:
			ip = v.IP
		}
		return ip
	}
	return nil
}

/*
Determine node ndame from IP. Works only in docker
*/
func getNodeName(ip net.IP) ID {
	ipArr := strings.Split(ip.String(), ".")

	a, _ := strconv.Atoi(ipArr[3])
	b := make([]byte, 20)
	binary.LittleEndian.PutUint64(b, uint64(a))
	var arr ID
	copy(arr[:ID_LENGTH], b[:])
	//log.Println(ip.String(), ipArr[3], b, arr)
	return arr
}

/*
Start server only
*/
func (node *Node) IStart() {
	go node.StartListening()
}

/*
Start server and try to connect to some other host.
*/
//TODO: FIX FOR TEST CASES
func (node *Node) Start(addr string) {
	go node.StartListening()
	timer1 := time.NewTimer(time.Duration(2) * time.Second)
	<-timer1.C

	go node.AddrSend(addr, JOIN_MSG, "Hello")
}
