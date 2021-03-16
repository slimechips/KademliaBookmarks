package node

import (
	"bufio"
	"container/list"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
)

type NodeCore struct {
	node_ID ID
	IP      net.IP
	Port    int
}

func (node NodeCore) String() string {
	return fmt.Sprintf("%d~%s~%d~;", node.node_ID.String(), node.IP.String(), node.Port)
}

type Node struct {
	node_core     NodeCore
	routing_table RoutingTable
	data          map[ID]string //stores a <key,value> pair for retrieval
	alive         bool          //for unit testing during prototyping
}

func (n *Node) Update(otherNodeCore *NodeCore) {
	prefix_length := otherNodeCore.node_ID.Xor(n.node_core.node_ID).PrefixLen()

	bucket := n.routing_table.Buckets[prefix_length]

	var element *list.Element = nil
	for e := bucket.Front(); e != nil; e = e.Next() {
		if e.Value.(NodeCore).node_ID.Equals((otherNodeCore).node_ID) {
			element = e
			break
		}
	}
	if element == nil {
		if bucket.Len() <= BUCKET_SIZE {
			bucket.PushFront(otherNodeCore)
		}
		// TODO: Handle insertion when the list is full by evicting old elements if
		// they don't respond to a ping.

	} else {
		bucket.MoveToFront(element)
	}
}

func (node Node) FullAddr() string {
	return fmt.Sprintf("%s:%d", node.node_core.IP.String(), node.node_core.Port)
}

func (n *Node) FindClosest(target ID, count int) (ret []NodeCore) {
	records := n.FindClosestRecord(target, count)
	for i := range records {
		ret = append(ret, *records[i].Node)
	}
	return
}

func (n *Node) FindClosestRecord(target ID, count int) (ret []nodeCoreRecord) {
	bucket_num := target.Xor(n.node_core.node_ID).PrefixLen()
	bucket := n.routing_table.Buckets[bucket_num]
	copyToList(bucket, &ret, target)

	// bidirectional search from the middle node
	for i := 1; (bucket_num-i >= 0 || bucket_num+i < ID_LENGTH*8) && len(ret) < count; i++ {
		if bucket_num-i >= 0 {
			bucket = n.routing_table.Buckets[bucket_num-i]
			copyToList(bucket, &ret, target)
		}
		if bucket_num+1 < ID_LENGTH*8 {
			bucket = n.routing_table.Buckets[bucket_num+i]
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

func createNodeCore(guid ID, ip_address net.IP, udp_port int) *NodeCore {
	n := NodeCore{node_ID: guid, IP: ip_address, Port: udp_port}
	return &n
}

//NewNode initalise node based on system IP
func NewNode(alive bool) *Node {
	ip := getIp()
	id := getNodeName(ip)
	node := &Node{
		node_core:     *createNodeCore(id, ip, RECEIVER_PORT),
		routing_table: *NewRoutingTable(),
		data:          make(map[ID]string),
		alive:         alive,
	}
	log.Println("Current Node Info:", *node)
	return node
}

func (node *Node) StartListening() {
	addr := net.UDPAddr{
		Port: node.node_core.Port,
		IP:   node.node_core.IP,
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

		if err != nil {

			log.Printf("Some error  %v\n", err)
			return
		}

		t := strings.Split(string(p), "|")
		if len(t) < 4 {
			log.Println("Error, received invalid message format")
			continue
		}
		senderID := convertStringToID(t[0])
		IP_address := net.ParseIP(t[1])
		tag := t[2]
		msgContent := t[3]

		switch tag {
		case JOIN_MSG:
			node.recvJoinReq(createNodeCore(senderID, IP_address, RECEIVER_PORT), ser, remoteAddr)
		case LIST_MSG:
			nodeList := strings.Split(msgContent, ";")
			node.recvJoinListNodeCore(convertStringToNodeCoreList(nodeList))
		case PING_MSG:
			node.recvPing(ser, remoteAddr)
		case STORE_MSG:
			kv := strings.Split(msgContent, ";")
			node.recvStore(convertStringToID(kv[0]), kv[1])
		case FVALUE_MSG:
			node.recvTryFindKeyValue(convertStringToID(msgContent), ser, remoteAddr)
		case FNODE_MSG:
			node.recvFindNode(convertStringToID(msgContent), ser, remoteAddr)
		// case "joinReq":
		// 	peerInfo := strings.Split(recvMsg, ";")
		// 	if len(peerInfo) < 2 {
		// 		log.Println("AddPeer format invalid")
		// 		continue
		// 	}
		// 	newPeerId, _ := strconv.Atoi(peerInfo[0])
		// 	newPeerIp := net.ParseIP(peerInfo[1])
		// 	log.Printf("Node %d-Recv Msg about NewPeer(Node id: %d, Ip: %s) from Node %d\n",
		// 		node.ID.String(), newPeerId, newPeerIp, senderID)
		// 	node.tryAddtoNetwork(newPeerId, newPeerIp)
		default: // e.g. hello
			// log.Printf("Node %d-Recv Msg from Node %d(%s): %s \n", node.ID.String(), senderID, senderIp.String(), recvMsg)
			// if node.tryAddtoNetwork(senderID, senderIp) {
			// 	node.broadcastNewPeer(senderID, senderIp)
			// }
		}
	}
}

func convertStringToID(s string) ID {
	i, _ := hex.DecodeString(s)
	var id_string ID
	copy(id_string[:ID_LENGTH], i)
	return id_string
}

func convertStringToNodeCoreList(stringList []string) []*NodeCore {
	nodecoreList := make([]*NodeCore, 0)
	for _, s := range stringList {
		s1 := strings.Split(s, "~")
		id := convertStringToID(s1[0])
		Port, _ := strconv.Atoi(s1[2])
		nodecoreList = append(nodecoreList,
			createNodeCore(id, net.ParseIP(s1[1]), Port))
	}
	return nodecoreList
}

//recvJoinReq appends new nodecore to my routing table and send listRecv to this new nodecore
func (node *Node) recvJoinReq(nodecore *NodeCore, conn *net.UDPConn, addr *net.UDPAddr) {
	prefix_length := nodecore.node_ID.Xor(node.node_core.node_ID).PrefixLen()
	go node.SendResponse(LIST_MSG, convertKBucketToString(node.routing_table.Buckets[prefix_length]), conn, addr)
	node.Update(nodecore)
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
	node.data[key] = value
}

func (node *Node) recvTryFindKeyValue(key ID, conn *net.UDPConn, addr *net.UDPAddr) {
	if val, ok := node.data[key]; ok {
		go node.SendResponse(FVALUE_MSG, val, conn, addr)
	} else {
		nodeCores := node.FindClosest(key, K)
		s := ""
		for _, nodeCore := range nodeCores {
			s += nodeCore.String() + ";"
		}
		go node.SendResponse(FVALUEFAIL_MSG, s, conn, addr)
	}

}
func (node *Node) recvFindNode(key ID, conn *net.UDPConn, addr *net.UDPAddr) {
	nodeCores := node.FindClosest(key, K)
	s := ""
	for _, nodeCore := range nodeCores {
		s += nodeCore.String() + ";"
	}
	go node.SendResponse(FVALUEFAIL_MSG, s, conn, addr)

}

func convertKBucketToString(bucket *list.List) string {
	s := ""
	for e := bucket.Front(); e != nil; e = e.Next() {
		NodeCore := e.Value.(*NodeCore)
		s += NodeCore.String()
	}
	return s
}

/*
format: <senderNodeID>|<IP_address>|<tag>|<msgContent>|
Tags: join, list, ping, store, fValue, fNode
*/
func (node *Node) Send(nodeCore *NodeCore, tag string, rawMsg string) {
	// Dont send to yourself
	if node.node_core.IP.String() == nodeCore.IP.String() {
		return
	}
	addr := fmt.Sprintf("%s:%d", nodeCore.IP.String(), nodeCore.Port)
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("Some error %v", err)
		return
	}

	msg := fmt.Sprintf("%s|%s|%s|%s|", node.node_core.node_ID.String(), node.node_core.IP.String(), tag, rawMsg)
	log.Printf("Node %d->Sending to %s:%s\n", node.node_core.node_ID, addr, msg)
	fmt.Fprintf(conn, msg)
	//waiting for response
	_, err = bufio.NewReader(conn).Read(p)
	ResponseMsgHandler(string(p))
	if err == nil {
		log.Printf("%s\n", p)
	} else {
		log.Printf("Some error %v\n", err)
	}
	conn.Close()
}

func ResponseMsgHandler(s string) {
	t := strings.Split(s, "|")
	if len(t) < 4 {
		//TODO: handle error better
		log.Println("Error, received invalid message format")
		return
	}
	//senderID := convertStringToID(t[0])
	//IP_address := net.ParseIP(t[1])
	tag := t[2]
	msgContent := t[3]
	switch tag {
	case PING_MSG:
		//TODO: pass (drop)
	case FVALUE_MSG:
		fmt.Println(msgContent)
	case FVALUEFAIL_MSG:
		//TODO: keep a list of nodes ive sent before & iterate through list of node cores FVALUE
	case FNODE_MSG:
		//TODO: send a STORE to the k closest to store the key-value
	}

}

func (node *Node) SendResponse(tag string, rawMsg string, conn *net.UDPConn, addr *net.UDPAddr) {
	msg := fmt.Sprintf("%s|%s|%s|%s|", node.node_core.node_ID.String(), node.node_core.IP.String(), tag, rawMsg)
	_, err := conn.WriteToUDP([]byte(msg), addr)
	if err != nil {
		log.Printf("Couldn't send response %v", err)
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
Determine node name from IP. Works only in docker
*/
func getNodeName(ip net.IP) ID {
	ipArr := strings.Split(ip.String(), ".")
	b, _ := hex.DecodeString(ipArr[3])
	var arr ID
	copy(arr[:ID_LENGTH], b[:])
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
	go node.Send(addr, "Hello", "Hello")
}
