package node

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const K = 2
const ID_LENGTH = 20
const JOIN_MSG = "join"
const LIST_MSG = "list"
const PING_MSG = "ping"
const STORE_MSG = "store"
const FVALUE_MSG = "fValue"
const FVALUEFAIL_MSG = "fValueF"
const FNODE_MSG = "fNode"

type ID [ID_LENGTH]byte

type routingTable struct {
	buckets [K][]*NodeCore
}

type Node struct {
	node_core     NodeCore
	routing_table routingTable
	data          map[ID]string //stores a <key,value> pair for retrieval
	alive         bool          //for unit testing during prototyping
}

type NodeCore struct {
	node_ID ID
	IP      net.IP
	Port    int
}

// Server listen port
const RECEIVER_PORT = 1053

func (id ID) String() string {
	return hex.EncodeToString(id)
}
func (node NodeCore) String() string {
	return fmt.Sprintf("%d~%s~%d~;", node.ID.String(), node.IP.String(), node.Port)
}

/*
Find out if node already exists, by ID
*/
func (node *Node) ContainsNode(id int) bool {
	if node.ID == id {
		log.Printf("%d-it's me\n", node.ID.String())
		return true
	} else {
		for _, a := range node.Peers {
			if a.ID == id {
				return true
			}
		}
	}
	return false
}

func (node Node) FullAddr() string {
	return fmt.Sprintf("%s:%d", node.IP, node.Port)
}

func createNodeCore(guid int, ip_address net.IP, udp_port int) *NodeCore {
	n := NodeCore{ID: guid, IP: ip_address, Port: udp_port}
	return &n
}

/*
Initalise node based on system IP
*/
func NewNode(alive bool) *Node {
	ip := getIp()
	id := getNodeName(ip)
	node := &Node{
		node_core:     *createNodeCore(id, ip, RECEIVER_PORT),
		routing_table: routingTable{buckets: make([K][]*NodeCore)},
		data:          make(map[int]int),
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
		senderID, _ := strconv.Atoi(t[0])
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
			node.recvStore(StringToFixedByte(kv[0]), kv[1])
		case FVALUE_MSG:
			node.recvTryFindKeyValue(StringToFixedByte(msgContent), ser, remoteAddr)
		case FNODE_MSG:
			node.recvFindNode(StringToFixedByte(msgContent), ser, remoteAddr)
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

func StringToFixedByte(s string) [ID_LENGTH]byte {
	var arr [ID_LENGTH]byte
	slice := []byte(s)
	fmt.Println(slice[:])
	copy(arr[:ID_LENGTH], slice[:])
	return arr
}

func convertStringToNodeCoreList(stringList []string) []*NodeCore {
	nodecoreList := make([]*NodeCore, 0)
	for _, s := range stringList {
		s1 := strings.Split(s, "~")
		ID, _ := strconv.Atoi(s1[0])
		Port, _ := strconv.Atoi(s1[2])
		nodecoreList = append(nodecoreList,
			createNodeCore(ID, net.ParseIP(s1[1]), Port))
	}
	return nodecoreList
}

//recvJoinReq appends new nodecore to my routing table and send listRecv to this new nodecore
func (node *Node) recvJoinReq(nodecore *NodeCore, conn *net.UDPConn, addr *net.UDPAddr) {
	//TODO: XOR function & edit bucket num to convertKBucketToString
	go node.SendResponse(LIST_MSG, convertKBucketToString(node.routing_table.buckets[nil]), conn, addr)
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

func (node *Node) recvStore(key [ID_LENGTH]byte, value string) {
	node.data[key] = value
}

func (node *Node) recvTryFindKeyValue(key [ID_LENGTH]byte, conn *net.UDPConn, addr *net.UDPAddr) {
	if val, ok := node.data[key]; ok {
		go node.SendResponse(FVALUE_MSG, val, conn, addr)
	} else {
		//TODO: XOR function
		nodeCores := node.FindClosest(K, key)
		//TODO: getKClosestNodeCores
		s := ""
		for _, nodeCore := range nodeCores {
			s += nodeCore.String() + ";"
		}
		go node.SendResponse(FVALUEFAIL_MSG, s, conn, addr)
	}

}
func (node *Node) recvFindNode(key [ID_LENGTH]byte, conn *net.UDPConn, addr *net.UDPAddr) {
	//TODO XOR
	nodeCores := node.FindClosest(K, key)
	s := ""
	for _, nodeCore := range nodeCores {
		s += nodeCore.String() + ";"
	}
	go node.SendResponse(FVALUEFAIL_MSG, s, conn, addr)

}

func (node *Node) tryAddtoNetwork(id int, ip net.IP) bool {
	if !node.ContainsNode(id) {
		log.Printf("Node %d-Added new peer in network. ID: %d, IP: %s\n", node.ID.String(), id, ip)
		node.Peers = append(node.Peers, Node{id, ip, RECEIVER_PORT, nil})
		return true
	}
	return false
}

func convertKBucketToString(nodeCores []*NodeCore) string {
	s := ""
	for _, nodeCore := range nodeCores {
		s += nodeCore.String()
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

	msg := fmt.Sprintf("%d|%s|%s|%s|", node.node_core.ID, node.node_core.IP.String(), tag, rawMsg)
	log.Printf("Node %d->Sending to %s:%s\n", node.node_core.ID, addr, msg)
	fmt.Fprintf(conn, msg)
	//waiting for response
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		log.Printf("%s\n", p)
	} else {
		log.Printf("Some error %v\n", err)
	}
	conn.Close()
}

func SendMsgHandler(s string) string {
	//1.alive
	//2.recv value from finding key
	//2.5 recv nodeCores from finding key (fail)
	//3.recv nodeCore to store key
	//
}

func (node *Node) SendResponse(tag string, rawMsg string, conn *net.UDPConn, addr *net.UDPAddr) {
	msg := fmt.Sprintf("%d|%s|%s|%s|", node.node_core.ID, node.node_core.IP.String(), tag, rawMsg)
	_, err := conn.WriteToUDP([]byte(msg), addr)
	if err != nil {
		log.Printf("Couldn't send response %v", err)
	}
}

/*
Broadcast a new peer to existing peers you know
*/
func (node *Node) broadcastNewPeer(id int, ip net.IP) {
	for _, peer := range node.Peers {
		log.Printf("Node %d-Sending to %s - New Peer Id: %d New Peer Ip: %s\n",
			node.ID.String(), peer.FullAddr(), id, ip.String())

		msg := fmt.Sprintf("%d;%s;", id, ip.String())
		go node.Send(peer.IP.String(), "AddPeer", msg)
	}
}

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
func getNodeName(ip net.IP) int {
	ipArr := strings.Split(ip.String(), ".")
	nodeName, _ := strconv.Atoi(ipArr[3])
	return nodeName
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
func (node *Node) Start(addr string) {
	go node.StartListening()
	go node.Send(addr, "Hello", "Hello")
}
