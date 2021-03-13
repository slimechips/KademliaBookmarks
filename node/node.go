package node

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type NodeAddr struct {
	IP   net.IP
	Port int
}

type Node struct {
	ID    int
	IP    net.IP
	Port  int
	Peers []NodeAddr
}

func contains(s []NodeAddr, e string) bool {
	for _, a := range s {
		fmt.Println(a.String(), e)
		if a.String() == e {
			return true
		}
	}
	return false
}

func SendResponse(msg string, conn *net.UDPConn, addr *net.UDPAddr) {
	_, err := conn.WriteToUDP([]byte(msg), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}

func ConvertToNodeAddr(addr string) NodeAddr {
	s := strings.Split(addr, ":")
	i, _ := strconv.Atoi(s[1])
	return NodeAddr{IP: net.IP(s[0]), Port: i}
}

func (node Node) String() string {
	return fmt.Sprintf("%s:%d", node.IP.String(), node.Port)
}

func (node Node) getAddr() string {
	return node.IP.String() + ":" + strconv.Itoa(node.Port)
}

func (nodeaddr NodeAddr) String() string {
	return nodeaddr.IP.String() + ":" + strconv.Itoa(nodeaddr.Port)
}

//newNode returns a new node
func NewNode(id int, ip string, port string) *Node {
	p, _ := strconv.Atoi(port)
	return &Node{
		ID:    id,
		IP:    net.ParseIP(ip),
		Port:  p,
		Peers: make([]NodeAddr, 0),
	}
}

func (node *Node) StartListening() {
	p := make([]byte, 2048)
	addr := net.UDPAddr{
		Port: node.Port,
		IP:   node.IP,
	}
	fmt.Printf("%v-Started listening\n", node.String())
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	ser.SetReadBuffer(1048576)
	for {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		if strings.Contains(string(p), "!") {
			s := strings.Split(string(p), "-")

			if contains(node.Peers, s[1]) {
				fmt.Printf("Added new peer in network\n")
				node.Peers = append(node.Peers, ConvertToNodeAddr(remoteaddr.String()))
			}
			fmt.Println("Passed")
		} else {
			s := strings.Split(string(p), "-")
			fmt.Printf("%v-Read a message from %v %s %v \n", node.String(), node.ID, remoteaddr, string(p))
			if contains(node.Peers, s[0]) {
				fmt.Printf("Aded new peer in network\n")
				node.Peers = append(node.Peers, ConvertToNodeAddr(s[0]))
			}
			if err != nil {

				fmt.Printf("Some error  %v\n", err)
				return
			}
			go SendResponse(node.String()+"-Hello I got your message \n", ser, remoteaddr)
			for _, peer := range node.Peers {
				conn, err := net.Dial("udp", peer.String())
				if err != nil {
					fmt.Printf("Some error %v\n", err)
					return
				}
				fmt.Printf("[Node %d]:[!]-"+peer.String()+"\n", node.ID)
				fmt.Fprintf(conn, "[!]-"+peer.String())
				conn.Close()
			}
		}
	}
}

func (node *Node) Contact(addr string) {
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	node.Peers = append(node.Peers, ConvertToNodeAddr(addr))
	fmt.Printf("%s:Sending to %s\n", node.String(), addr)
	fmt.Fprintf(conn, node.String()+"-Hello, How are you doing?\n")
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		fmt.Printf("%s\n", p)
	} else {
		fmt.Printf("Some error %v\n", err)
	}
	conn.Close()
}

func (node *Node) IStart() {
	go node.StartListening()
}

func (node *Node) Start(addr string) {
	go node.Contact(addr)
	go node.StartListening()
}
