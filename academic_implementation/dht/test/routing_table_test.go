package main

import (
	"KademliaBookmarks/academic_implementation/dht/kademlia"
	"testing"
)

func TestRoutingTable(t *testing.T) {
	// currently the node is just using hexadecimal conversion so that it is easy to tell which node closer to which
	// later will change it to SHA1 to encode the hostname
	currentNode := kademlia.NewID("0000000000000000000000000000000000000000")
	n1 := kademlia.NewID("FFFFFFFF00000000000000000000000000000000")
	n2 := kademlia.NewID("FFFFFFF000000000000000000000000000000000")
	n3 := kademlia.NewID("1111111100000000000000000000000000000000")

	// create new routing table for current node
	rt := kademlia.NewRoutingTable(&kademlia.Contact{currentNode, "localhost:8000"})
	rt.Update(&kademlia.Contact{n1, "localhost:8001"})
	rt.Update(&kademlia.Contact{n2, "localhost:8002"})
	rt.Update(&kademlia.Contact{n3, "localhost:8003"})

	list := rt.FindClosest(kademlia.NewID("2222222200000000000000000000000000000000"), 1)
	if len(list) != 1 {
		t.Fail()
		return
	}
	if !list[0].Node.Id.Equals(n3) {
		t.Error(list[0])
	}

	list = rt.FindClosest(n2, 2)
	if len(list) != 2 {
		t.Error(len(list))
		return
	}
	if !list[0].Node.Id.Equals(n2) {
		t.Error(list[0])
	}
	if !list[1].Node.Id.Equals(n1) {
		t.Error(list[1])
	}

	// add more to the routing table
	n4 := kademlia.NewID("0111111100000000000000000000000000000000")
	n5 := kademlia.NewID("0F11111100000000000000000000000000000000")
	rt.Update(&kademlia.Contact{n4, "localhost:8001"})
	rt.Update(&kademlia.Contact{n5, "localhost:8002"})

	list = rt.FindClosest(kademlia.NewID("0211111100000000000000000000000000000000"), 5)
	if !list[0].Node.Id.Equals(n4) {
		t.Error(list[0])
	}
}

func TestRoutingTableWithSHA1(t *testing.T) {
	rt := kademlia.NewRoutingTable(kademlia.NewContact("localhost:8000"))
	rt.Update(kademlia.NewContact("localhost:8001"))
	rt.Update(kademlia.NewContact("localhost:8002"))
	rt.Update(kademlia.NewContact("localhost:8003"))
}
