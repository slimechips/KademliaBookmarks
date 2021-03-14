package main

import (
	"log"
	"os"
	"runtime"

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
	node0 := N.NewNode()
	go node0.Start(subnetStart + "1")

	runtime.Goexit()
	log.Println("Exiting")

	// <-time.NewTimer(time.Duration(time.Second * 10)).C
	// fmt.Println(node0.Peers)
	// fmt.Println(node1.Peers)
	// fmt.Println(node2.Peers)

}
