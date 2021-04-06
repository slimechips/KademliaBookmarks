package node

import (
	"fmt"
	"time"
)

type Node struct {
	Id      int
	Peer    *Node
	Data    map[int]*Item //stores a <key,value> pair for retrieval
	Alive   chan bool     //for unit testing during prototyping
	Publish chan bool
}

func NewNode(Id int) *Node {
	return &Node{
		Id:      Id,
		Data:    map[int]*Item{},
		Peer:    nil,
		Alive:   make(chan bool),
		Publish: make(chan bool),
	}
}

func (node *Node) Store(other *Node, key int, item *Item) {
	other.Data[key] = NewItem(item.Value)
}

func (node *Node) Republish(timing time.Time) {
	for k, v := range node.Data {
		if v.IsTimeToPublish(timing) {
			// republish by running store functions
			fmt.Println("Send republished")

			// in this case just republish itself
			node.Data[k] = NewItem(v.Value)
		}
	}
}

func (node *Node) String() string {
	return fmt.Sprintf("Id: %d\n", node.Id) + fmt.Sprint("data: ", node.Data)
}

func (node *Node) PeriodicPublish() {
	go func(publish chan bool) {
		fmt.Println(REPUBLISHED_DURATION)
		<-time.NewTimer(REPUBLISHED_DURATION).C
		publish <- true
	}(node.Publish)
}

func (node *Node) Listen() {
	// running periodic functions
	node.PeriodicPublish()
Running:
	for {
		select {
		case <-node.Publish:
			// by right it should be current time
			node.Republish(time.Now())

			// print
			fmt.Println(node)

			// run another periodic routine
			node.PeriodicPublish()
		case <-node.Alive:
			fmt.Println("DEAD BOY")
			break Running
		}
	}
}
