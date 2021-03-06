package main

import (
	"bufio"
	"container/list"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Node struct {
	NodeCore     NodeCore
	RoutingTable RoutingTable
	Data         map[ID]*Item //stores a <key,value> pair for retrieval
	Alive        bool         //for unit testing during prototyping
	PublishChan  chan bool
	ExpiryChan   chan bool
	mutex        *sync.Mutex
	jMutex       *sync.Mutex
	Cache        []string
}

func (node *Node) Init(nodeID, targetIP string) {
	node.NodeCore.GUID = NewSHA1ID(nodeID)
}

func (node *Node) addToCache(key string) {

	for _, b := range node.Cache {
		if b == key {
			return
		}
	}
	node.Cache = append(node.Cache, key)

}

func (node *Node) getCacheOnlyKeys() []string {
	keysOnly := make([]string, 0)
	for _, b := range node.Cache {
		if !strings.Contains(b, "!") {
			if _, ok := node.Data[NewSHA1ID(b)]; ok {
				keysOnly = append(keysOnly, b)
			}
		}
	}
	return keysOnly
}
func (node *Node) Republish() {
	for key, item := range node.Data {
		log.Printf("I have key in data: <%s,%s>\n", key.String(), item)
		// check if you are supposed to republish
		// assume it is supposed to store
		if item.IsTimeToPublish(time.Now()) {
			// restore the nodes
			log.Printf("I am republishing <%s,%s>\n", key, item.Value)
			nodeCores := node.KNodesLookUp(key)
			for i := range nodeCores {
				log.Printf("I am sending to %s\n", nodeCores[i])
			}
			value := item.Value
			node.StoreInNodes(nodeCores, key, value)
		}
	}
}

func (node *Node) CheckExpiredData() {
	for key, item := range node.Data {
		// check if you are supposed to republish
		// assume it is supposed to store
		if item.IsItExpired(time.Now()) {
			// delete the data
			delete(node.Data, key)
		}
	}
}
func (n *Node) Update(otherNodeCore *NodeCore) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	prefix_length := otherNodeCore.GUID.Xor(n.NodeCore.GUID).PrefixLen()

	bucket := n.RoutingTable.Buckets[prefix_length]

	var element *list.Element = nil
	for e := bucket.Front(); e != nil; e = e.Next() {
		if e.Value.(*NodeCore).GUID.Equals((otherNodeCore).GUID) {
			element = e
			break
		}
	}
	if element == nil {
		if bucket.Len() <= K {
			log.Printf("Pushing to %d bucket: Node %s\n", prefix_length, otherNodeCore.IP.String())
			bucket.PushFront(otherNodeCore)
		} else {
			// TODO: Handle insertion when the list is full by evicting old elements if
			log.Println(bucket)
			LRUNode := bucket.Back()
			pingok := make(chan string)
			go n.Send(LRUNode.Value.(*NodeCore), PING_MSG, "hi", pingok)
			timer := time.NewTimer(TIMEOUT_DURATION)
			select {
			case <-pingok:
				timer.Stop()
				log.Println(LRUNode.Value.(*NodeCore).String() + "is alive")
				log.Printf("Pushing to %d bucket: Node %s\n", prefix_length, LRUNode.Value.(*NodeCore).IP.String())
				bucket.MoveToFront(LRUNode)
			case <-timer.C:
				// remove the LRUNode
				bucket.Remove(LRUNode)
				// add the new node to the front
				log.Printf("Pushing to %d bucket: Node %s\n", prefix_length, otherNodeCore.IP.String())
				bucket.PushFront(otherNodeCore)
			}
		}
	} else {
		bucket.MoveToFront(element)
	}
	// s := "MY RT: "
	// for i := 0; i < ID_LENGTH*8; i++ {
	// 	bckt := n.RoutingTable.Buckets[i]
	// 	for e := bckt.Front(); e != nil; e = e.Next() {
	// 		s += e.Value.(*NodeCore).String()
	// 	}
	// }
	// log.Println(s)
	// n.mutex.Unlock()

}

func (n *Node) getData() []string {
	temp := make([]string, 0)
	for _, v := range n.Data {
		temp = append(temp, v.Value)
	}
	return temp
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

func (n *Node) checkAlive(other *NodeCore) bool {
	pingok := make(chan string)
	go n.Send(other, PING_MSG, "hi", pingok)
	timer := time.NewTimer(TIMEOUT_DURATION)
	select {
	case <-pingok:
		timer.Stop()
		log.Println(other.String() + "is alive")
		return true
	case <-timer.C:
		return false
	}
}
func (n *Node) FindClosestRecord(target ID, count int) (ret []nodeCoreRecord) {
	bucket_num := target.Xor(n.NodeCore.GUID).PrefixLen()
	bucket := n.RoutingTable.Buckets[bucket_num]
	copyToList(bucket, &ret, target)

	// bidirectional search from the middle node
	for i := 1; (bucket_num-i >= 0 || bucket_num+i < ID_LENGTH*8) && len(ret) < count; i++ {
		if bucket_num-i >= 0 {
			bucket = n.RoutingTable.Buckets[bucket_num-i]
			for e := bucket.Front(); e != nil; e = e.Next() {
				NodeCore := e.Value.(*NodeCore)
				if n.checkAlive(NodeCore) {
					addToList(NodeCore, &ret, target)
				}
			}
		}
		if bucket_num+1 < ID_LENGTH*8 {
			bucket = n.RoutingTable.Buckets[bucket_num+i]
			for e := bucket.Front(); e != nil; e = e.Next() {
				NodeCore := e.Value.(*NodeCore)
				if n.checkAlive(NodeCore) {
					addToList(NodeCore, &ret, target)
				}
			}
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
	node.jMutex.Lock()
	log.Println("Locking JMutex")
	defer func() {
		node.jMutex.Unlock()
		log.Println("Unlocking JMutex")
	}()
	nodeCores := node.FindClosest(nodecore.GUID, K-1)
	s := node.NodeCore.String()
	for _, nc := range nodeCores {
		s += nc.String()
	}
	log.Println("JOINREQ: " + s)
	go node.SendResponse(LIST_MSG, s, conn, addr)
	node.Update(nodecore)

}

func (node *Node) recvJoinListNodeCore(list []*NodeCore) {
	// node.Update(sentNodeCore)
	log.Printf("LIST:%v\n", list)
	if list == nil {
		return
	}
	for _, nodeCore := range list {
		pingOK := make(chan string)
		log.Println("pinging: " + nodeCore.String())
		go node.Send(nodeCore, PING_MSG, "hi", pingOK)
		timer := time.NewTimer(TIMEOUT_DURATION)
		select {
		case <-pingOK:
			timer.Stop()
			log.Println(nodeCore.String() + "is alive")
			node.Update(nodeCore)
		case <-timer.C:
			log.Println("PING TIMEOUT")
		}
	}
}

func (node *Node) recvPing(nodecore *NodeCore, conn *net.UDPConn, addr *net.UDPAddr) {
	go node.SendResponse(PING_MSG, "hi", conn, addr)
	log.Println("received Ping from:" + nodecore.String())
	node.Update(nodecore)
}

func (node *Node) recvStore(key ID, value string) {
	node.Data[key] = NewItem(value)
	log.Println("VALUECACHE:" + value)
	s := strings.Split(value, "*")
	if len(s) > 0 {
		node.addToCache(s[0])
	}

}

func (node *Node) recvTryFindKeyValue(key ID, conn *net.UDPConn, addr *net.UDPAddr) {
	if item, ok := node.Data[key]; ok {
		go node.SendResponse(FVALUE_MSG, item.Value, conn, addr)
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
			log.Println("LOOKUP_TIMEOUT")
			break iterativeFind
		case msg := <-chanFail:
			timer = time.NewTimer(time.Duration(TIMEOUT_DURATION))
			log.Println("MSGTOLOOKUP:" + msg)
			s := strings.Split(msg, "#")

			if !StringsListContains(s[0], alive) {
				alive = append(alive, s[0])
			}
			ss := strings.Split(s[1], ";")
			if len(ss) >= 2 {
				nodeCoreList := convertStringToNodeCoreList(ss)
				for _, n := range nodeCoreList {
					if !StringsListContains(n.GUID.String(), requested) {
						requested = append(requested, n.GUID.String())
						node.Update(n)
						go node.Send(n, FLOOKUP_MSG, key.String(), chanFail, chanSucc)
					}
				}
			}

		case <-chanSucc:
			timer.Stop()
			log.Println("LOOKUP_SUCCESS")
			break iterativeFind
		}
	}
	alive = append(alive, node.NodeCore.String())
	log.Printf("ALIVE: %v", alive)
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
	log.Println("STORE_IN_NODES,", key.String(), val)
	for _, nodeCore := range nodeCores {
		if nodeCore.GUID.Equals(node.NodeCore.GUID) {
			go node.recvStore(key, val)
		} else {

			go node.Send(nodeCore, STORE_MSG, fmt.Sprintf("%s;%s;", key.String(), val), nil)
		}
	}
}

//FindValueByKey sends request to Nodes for target key-value
func (node *Node) FindValueByKey(key ID) string {
	if val, ok := node.Data[key]; ok {
		log.Println(val.Value)
		return val.Value
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
			ss := strings.Split(s[1], ";")
			if len(ss) >= 2 {
				nodeCoreList := convertStringToNodeCoreList(ss)
				for _, n := range nodeCoreList {
					if !StringsListContains(n.GUID.String(), requested) {
						requested = append(requested, n.GUID.String())
						node.Update(n)
						go node.Send(n, FVALUE_MSG, key.String(), chanFail, chanSucc)
					}
				}
			}
		case msg := <-chanSucc:
			timer.Stop()
			return strings.Split(msg, "#")[1]
		}
	}

	return "value not found"
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
func (node *Node) AddrSend(addr string, tag string, rawMsg string, opt ...chan string) {
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
	node.ResponseMsgHandler(fmt.Sprintf("~%s~%s~", ad[0], ad[1]), string(p), opt)
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

func createNodeCoreNoGUID(ip_address net.IP, udp_port int) *NodeCore {
	n := NodeCore{IP: ip_address, Port: udp_port}
	return &n
}

//NewNode initalise node based on system IP
func NewNode(alive bool, args []string) *Node {
	ip := getIp()
	// id := getNodeName(ip)
	node := &Node{
		NodeCore:     *createNodeCoreNoGUID(ip, RECEIVER_PORT),
		RoutingTable: *NewRoutingTable(),
		Data:         make(map[ID]*Item),
		PublishChan:  make(chan bool),
		ExpiryChan:   make(chan bool),
		Alive:        alive,
		mutex:        &sync.Mutex{},
		jMutex:       &sync.Mutex{},
		Cache:        make([]string, 0),
	}
	if len(args) <= 1 {
		node.NodeCore.GUID = getNodeName(ip)
	} else {
		node.Init(args[0], args[1])
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

	// Run all periodic checking
	go RepublishMessageNewsFlash(node.PublishChan)
	go DeleteDataIfExpireNewsFlash(node.ExpiryChan)

	// listening
	go func(node *Node) {
		for {
			select {
			case <-node.PublishChan:
				// republish
				log.Println("check republish")
				node.Republish()

				// periodic running the publish
				go RepublishMessageNewsFlash(node.PublishChan)
			case <-node.ExpiryChan:
				// Delete expired data
				node.CheckExpiredData()

				// periodic running the expiry
				go DeleteDataIfExpireNewsFlash(node.ExpiryChan)
			default:
				//TODO: IS THIS RIGHT? WE NEED TO TEST REPUBLISHING
			}
		}
	}(node)

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
		senderID := ConvertStringByteRepresentToID(t[0])
		IP_address := net.ParseIP(t[1])
		tag := t[2]
		msgContent := t[3]
		log.Println("RECEIVING:", msgContent)

		// Receiver perspective of what request send it
		switch tag {
		case JOIN_MSG:
			go node.recvJoinReq(createNodeCore(senderID, IP_address, RECEIVER_PORT), ser, remoteAddr)
		case PING_MSG:
			go node.recvPing(createNodeCore(senderID, IP_address, RECEIVER_PORT), ser, remoteAddr)
		case STORE_MSG:
			kv := strings.Split(msgContent, ";")
			go node.recvStore(ConvertStringByteRepresentToID(kv[0]), kv[1])
		case FVALUE_MSG:
			go node.recvTryFindKeyValue(ConvertStringByteRepresentToID(msgContent), ser, remoteAddr)
		case FLOOKUP_MSG:
			go node.recvFindNode(ConvertStringByteRepresentToID(msgContent), ser, remoteAddr)
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

// func ConvertStringToID(s string) ID {
// 	i, _ := hex.DecodeString(hex.EncodeToString([]byte(s)))
// 	var id_string ID
// 	//log.Println(i)
// 	copy(id_string[:ID_LENGTH], i)
// 	return id_string
// }

func ConvertStringByteRepresentToID(s string) ID {
	return NewID(s)
}

func StringsListContains(s string, stringList []string) bool {
	for _, str := range stringList {
		if s == str {
			return true
		}
	}
	return false
}

func convertStringToNodeCore(s string) *NodeCore {
	s1 := strings.Split(s, "~")
	id := ConvertStringByteRepresentToID(s1[0])
	Port, _ := strconv.Atoi(s1[2])
	return createNodeCore(id, net.ParseIP(s1[1]), Port)
}

func convertStringToNodeCoreList(stringList []string) []*NodeCore {
	nodecoreList := make([]*NodeCore, 0)
	for i := 0; i < len(stringList); i++ {
		//log.Println("CONVERTING TO NODECORE", stringList[i])
		s1 := strings.Split(stringList[i], "~")
		log.Println("STRINGLIST:" + stringList[i])
		if len(s1) > 2 {
			id := ConvertStringByteRepresentToID(s1[0])
			Port, _ := strconv.Atoi(s1[2])
			nodecoreList = append(nodecoreList,
				createNodeCore(id, net.ParseIP(s1[1]), Port))
		}
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
	// senderID := ConvertStringToID(t[0])
	senderID := ConvertStringByteRepresentToID(t[0])
	//IP_address := net.ParseIP(t[1])
	tag := t[2]
	msgContent := t[3]
	log.Println("RESPONSE_HANDLER:", senderID, tag, recvAddr, msgContent)
	switch tag {
	case LIST_MSG:
		chans[0] <- senderID.String() + recvAddr + "#" + msgContent
	case PING_MSG:
		chans[0] <- senderID.String() + recvAddr + "#" + msgContent
	case FVALUE_MSG:
		chans[1] <- senderID.String() + recvAddr + "#" + msgContent
		log.Println(msgContent)
	case FVALUEFAIL_MSG:
		chans[0] <- senderID.String() + recvAddr + "#" + msgContent
	case FLOOKUP_MSG:
		chans[0] <- senderID.String() + recvAddr + "#" + msgContent
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

	//log.Println(ip.String(), ipArr[3], b, arr)
	return NewSHA1ID(ipArr[3])
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
	// timer1 := time.NewTimer(time.Duration(2) * time.Second)
	// <-timer1.C

	res := make(chan string)
	go node.AddrSend(addr, JOIN_MSG, "Hello", res)
	timer := time.NewTimer(JOIN_WAIT_DURATION)
	select {
	case msg := <-res:
		timer.Stop()
		log.Println("Received Reply from recvjoinlist")
		s := strings.Split(msg, "#")
		msgContent := s[1]
		nodeList := strings.Split(msgContent, ";")

		node.recvJoinListNodeCore(convertStringToNodeCoreList(nodeList))
	case <-timer.C:
		log.Println("JOIN_LIST_TIMEOUT")

	}
}
