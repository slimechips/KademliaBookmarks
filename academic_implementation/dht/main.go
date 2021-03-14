package main

import (
	"KademliaBookmarks/academic_implementation/dht/kademlia"
	"fmt"
)

func main() {
	a := kademlia.ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	// b := simpleKademlia.ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 19, 18}
	// c := simpleKademlia.ID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	str_id := "0123456789abcdef0123456789abcdef01234567"
	str_id2 := "0123456789abcdef0123456789abcdef01234567"

	id1 := kademlia.NewID(str_id)
	id2 := kademlia.NewID(str_id2)

	fmt.Println(id1)
	fmt.Println(id1.Xor(id2))
	fmt.Println(id1.Xor(id2).PrefixLen())
	for _, n := range a {
		fmt.Printf("% 08b", n) // prints 00000000 11111101
	}

	n1 := kademlia.NewID("FFFFFFFF00000000000000000000000000000000")
	n2 := kademlia.NewID("FFFFFFF000000000000000000000000000000000")
	n3 := kademlia.NewID("1111111100000000000000000000000000000000")

	rt := kademlia.NewRoutingTable(&kademlia.Contact{n1, "localhost:8000"})
	rt.Update(&kademlia.Contact{n2, "localhost:8001"})
	rt.Update(&kademlia.Contact{n3, "localhost:8002"})
}
