package main

import (
	"KademliaBookmarks/academic_implementation/dht/kademlia"
	"fmt"
	"testing"
)

func TestID(t *testing.T) {
	a := kademlia.ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	b := kademlia.ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 19, 18}
	c := kademlia.ID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}

	if !a.Equals(a) {
		t.Fail()
	}
	if a.Equals(b) {
		t.Fail()
	}

	if !a.Xor(b).Equals(c) {
		t.Error(a.Xor(b))
	}

	if c.PrefixLen() != 151 {
		t.Error(c.PrefixLen())
	}

	if b.Less(a) {
		t.Fail()
	}

	str_id := "0123456789abcdef0123456789abcdef01234567"
	str_id2 := "0123456789abcdef13234f6789abcdef01234567"
	if kademlia.NewID(str_id).String() != str_id {
		t.Error(kademlia.NewID(str_id).String())
	}

	id1 := kademlia.NewID(str_id)
	id2 := kademlia.NewID(str_id2)
	fmt.Println(id1.Xor(id2).PrefixLen())
}
