package node

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type Node struct {
	ID    int
	IP    net.IP
	Port  int
	Peers []Node // Only current node will have this
}

// Server listen port
const RECEIVER_PORT = 1053

/*
Find out if node already exists, by ID
*/
func (node *Node) ContainsNode(id int) bool {
	if node.ID == id {
		log.Printf("%d-it's me\n", node.ID)
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

/*
Initalise node based on system IP
*/
func NewNode() *Node {
	ip := getIp()
	id := getNodeName(ip)
	node := &Node{
		ID:    id,
		IP:    ip,
		Port:  RECEIVER_PORT,
		Peers: make([]Node, 0),
	}
	log.Println("Current Node Info:", *node)
	return node
}

func (node *Node) StartListening() {
	addr := net.UDPAddr{
		Port: node.Port,
		IP:   node.IP,
	}
	log.Printf("%v-Started listening\n", node.FullAddr())
	// Listen on localhost
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Some error %v\n", err)
		return
	}
	for {
		// Byte array to hold message
		p := make([]byte, 2048)
		_, remoteaddr, err := ser.ReadFromUDP(p)

		if err != nil {

			log.Printf("Some error  %v\n", err)
			return
		}

		t := strings.Split(string(p), "|")
		if len(t) < 3 {
			log.Println("Error, received invalid message format")
			continue
		}
		senderID, _ := strconv.Atoi(t[0])
		tag := t[1]
		recvMsg := t[2]
		senderIp := remoteaddr.IP

		switch tag {
		case "AddPeer":
			peerInfo := strings.Split(recvMsg, ";")
			if len(peerInfo) < 2 {
				log.Println("AddPeer format invalid")
				continue
			}
			newPeerId, _ := strconv.Atoi(peerInfo[0])
			newPeerIp := net.ParseIP(peerInfo[1])
			log.Printf("Node %d-Recv Msg about NewPeer(Node id: %d, Ip: %s) from Node %d\n",
				node.ID, newPeerId, newPeerIp, senderID)
			node.tryAddtoNetwork(newPeerId, newPeerIp)
		default: // e.g. hello
			log.Printf("Node %d-Recv Msg from Node %d(%s): %s \n", node.ID, senderID, senderIp.String(), recvMsg)
			if node.tryAddtoNetwork(senderID, senderIp) {
				node.broadcastNewPeer(senderID, senderIp)
			}
		}
	}
}

func (node *Node) tryAddtoNetwork(id int, ip net.IP) bool {
	if !node.ContainsNode(id) {
		log.Printf("Node %d-Added new peer in network. ID: %d, IP: %s\n", node.ID, id, ip)
		node.Peers = append(node.Peers, Node{id, ip, RECEIVER_PORT, nil})
		return true
	}
	return false
}

/*
Send message to given address
Port suffix is added automatically
Message format: <NODE_ID>|<TAG>|<MSG>
Tags: Hello, AddPeer
*/
func (node *Node) Send(addr string, tag string, rawMsg string) {
	// Dont send to yourself
	if node.IP.String() == addr {
		return
	}
	addr = fmt.Sprintf("%s:%d", addr, RECEIVER_PORT)
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("Some error %v", err)
		return
	}

	msg := fmt.Sprintf("%d|%s|%s|", node.ID, tag, rawMsg)
	log.Printf("Node %d-Sending to %s:%s\n", node.ID, addr, msg)
	fmt.Fprintf(conn, msg)
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		log.Printf("%s\n", p)
	} else {
		log.Printf("Some error %v\n", err)
	}
	conn.Close()
}

func (node *Node) SendResponse(msg string, conn *net.UDPConn, addr *net.UDPAddr) {
	msg = fmt.Sprintf("%s|%s|", node.FullAddr(), msg)
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
			node.ID, peer.FullAddr(), id, ip.String())

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
