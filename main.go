package main

import (
	"log"
	"os"
	"runtime"
	"time"

	N "./node"
)

const subnetStart = "172.16.238."

func main() {

	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
	log.Println("Log file started")
	node0 := N.NewNode(true)
	go node0.Start(subnetStart + "1")
	<-time.NewTimer(time.Duration(10) * time.Second).C

	if node0.NodeCore.IP.String() == "172.16.238.1" {
		log.Println("We're node 1. Starting test case by storing")
		<-time.NewTimer(time.Duration(10) * time.Second).C
		key := N.ConvertStringToID("123")
		nodeCores := node0.KNodesLookUp(key)
		log.Printf("Node Cores: %d\n", len(nodeCores))
		node0.StoreInNodes(nodeCores, key, "Prof Sudipta rocks")
		<-time.NewTimer(time.Duration(10) * time.Second).C
		log.Println("Finding value by key now")
		node0.FindValueByKey(key)
	}

	runtime.Goexit()
	log.Println("Exiting")

	// <-time.NewTimer(time.Duration(time.Second * 10)).C
	// fmt.Println(node0.Peers)
	// fmt.Println(node1.Peers)
	// fmt.Println(node2.Peers)

}
