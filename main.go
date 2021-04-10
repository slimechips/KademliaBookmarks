package main

import (
	"log"
	"os"
	"runtime"
	"time"
)

const subnetStart = "10.0.0."
const logsDir = "./logs"

func main() {
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		os.Mkdir(logsDir, os.ModeDir)
	}
	file, err := os.OpenFile(logsDir+"/app.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
	log.Println("Log file started")
	node0 := NewNode(true, os.Args[1:])
	if len(os.Args) <= 1 {
		<-time.NewTimer(TIMEOUT_DURATION * 2).C
		go node0.Start(subnetStart + "1")
		<-time.NewTimer(time.Duration(10) * time.Second).C
	} else {
		log.Println("Got Args yo")
		go node0.Start(os.Args[2])
	}

	// if node0.NodeCore.IP.String() == "172.16.238.1" {
	// 	log.Println("We're node 1. Starting test case by storing")
	// 	<-time.NewTimer(time.Duration(10) * time.Second).C
	// 	key := ConvertStringToID("123")
	// 	nodeCores := node0.KNodesLookUp(key)
	// 	log.Printf("Node Cores: %d\n", len(nodeCores))
	// 	node0.StoreInNodes(nodeCores, key, "Prof Sudipta rocks")
	// 	<-time.NewTimer(time.Duration(10) * time.Second).C
	// 	log.Println("Finding value by key now")
	// 	node0.FindValueByKey(key)
	// }
	wb := initServer(node0)
	wb.runWebServer(8080)

	runtime.Goexit()
	log.Println("Exiting")

}
