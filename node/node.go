package node

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type NodeAddr struct {
	IP   string
	Port int
}

type Node struct {
	ID    int
	IP    string
	Port  int
	Peers []NodeAddr
}

func contains(s *[]NodeAddr, e string) bool {
	for _, a := range *s {
		fmt.Println(a.String(), e)
		if strings.Compare(a.String(), e) == 0 {
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
	i, _ := strconv.Atoi(strings.TrimSpace(s[1]))
	fmt.Println(s, s[1], i)
	return NodeAddr{IP: s[0], Port: i}
}

func (node Node) String() string {
	return fmt.Sprintf("%s:%d", node.IP, node.Port)
}

func (node Node) getAddr() string {
	return node.IP + ":" + strconv.Itoa(node.Port)
}

func (nodeaddr NodeAddr) String() string {
	return nodeaddr.IP + ":" + strconv.Itoa(nodeaddr.Port)
}

//newNode returns a new node
func NewNode(id int, ip string, port string) *Node {
	p, _ := strconv.Atoi(port)
	return &Node{
		ID:    id,
		IP:    ip,
		Port:  p,
		Peers: make([]NodeAddr, 0),
	}
}

func (node *Node) StartListening() {
	addr := net.UDPAddr{
		Port: node.Port,
		IP:   net.ParseIP(node.IP),
	}
	fmt.Printf("%v-Started listening\n", node.String())
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	for {
		p := make([]byte, 2048)
		_, remoteaddr, err := ser.ReadFromUDP(p)
		if err != nil {

			fmt.Printf("Some error  %v\n", err)
			return
		}
		if strings.Contains(string(p), "!") {

			s := strings.Split(string(p), "-")
			fmt.Println("Shared", s[0], s[1])
			if !contains(&node.Peers, s[1]) {
				fmt.Printf("Added new peer in network - %s\n", s[1])
				node.Peers = append(node.Peers, ConvertToNodeAddr(s[1]))
			}
			fmt.Println(s[1])
		} else {
			s := strings.Split(string(p), "-")
			fmt.Printf("%v-Read a message from %v %s %v \n", node.String(), node.ID, remoteaddr, string(p))
			fmt.Println(contains(&node.Peers, s[0]))
			if !contains(&node.Peers, s[0]) {
				fmt.Printf("Added new peer in network\n")
				node.Peers = append(node.Peers, ConvertToNodeAddr(s[0]))
			}
			go SendResponse(node.String()+"-Hello I got your message \n", ser, remoteaddr)
			for _, peer := range node.Peers {
				fmt.Printf("sending to %s - %s\n", peer.String(), s[0])
				conn, err := net.Dial("udp", peer.String())
				if err != nil {
					fmt.Printf("Some error %v\n", err)
					return
				}
				fmt.Printf("%s>[!]-"+peer.String()+">%s\n", node.String(), s[0])
				fmt.Fprintf(conn, "[!]-"+s[0])
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
	fmt.Fprintf(conn, node.String()+"-Hello\n")
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
