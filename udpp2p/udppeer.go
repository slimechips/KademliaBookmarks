package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)


func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	_, err := conn.WriteToUDP([]byte("From"), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}

// keep a copy
//build a node structure
func send(port int) {
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		fmt.Printf("%s\n", p)
	} else {
		fmt.Printf("Some error %v\n", err)
	}
	conn.Close()
}

func listen(port int) {
	p := make([]byte, 2048)
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("127.0.0.1"),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}
	for {
		_, remoteaddr, err := ser.ReadFromUDP(p)
		fmt.Printf("Read a message from %v %s \n", remoteaddr, p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		go sendResponse(ser, remoteaddr)
	}
}

func main() {
	fmt.Println("Please enter port for your peer")
	fmt.Println("---------------------")
	var o int
	fmt.Scan(&o)
	fmt.Println("Please enter port for yourself")
	fmt.Println("---------------------")
	var i int
	fmt.Scan(&i)
	send(o)
	listen(i)
}
