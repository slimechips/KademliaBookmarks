package main

import (
	"fmt"
	"time"

	N "./node"
)

func main() {
	node0 := N.NewNode(0, "127.0.0.1")
	node1 := N.NewNode(1, "127.0.0.2")
	node2 := N.NewNode(2, "127.0.0.3")
	go node0.IStart()
	go node1.Start("127.0.0.1:1234")
	go node2.Start("127.0.0.2:1234")

	<-time.NewTimer(time.Duration(time.Second * 10)).C
	fmt.Println(node0.Peers)
	fmt.Println(node1.Peers)
	fmt.Println(node2.Peers)

}
