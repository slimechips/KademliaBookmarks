package main

import (
	"KademliaBookmarks/academic_implementation/dht/kademlia"
	"crypto/sha1"
	"encoding/hex"
	"testing"
)

func TestID(t *testing.T) {
	// 20 byte ID
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
}

func TestPrefixLen(t *testing.T) {
	a := kademlia.ID{128, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	b := kademlia.ID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	c := kademlia.ID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	// differ in the last bit
	if c.Xor(b).PrefixLen() != 159 {
		t.Error(c.Xor(b))
		t.Errorf("Expected %v but got %v", 159, c.Xor(b).PrefixLen())
	}

	// differ in the 1st bit
	if a.Xor(b).PrefixLen() != 0 {
		t.Error(a.Xor(b))
		t.Errorf("Expected %v but got %v", 0, a.Xor(b).PrefixLen())
	}
}

func TestSHA1Hash(t *testing.T) {
	address := "192.128.0.0:1234"
	a := kademlia.NewSHA1ID(address)

	h := sha1.New()
	h.Write([]byte(address))
	hashed := h.Sum(nil)

	// compare
	for i := range a {
		if a[i] != hashed[i] {
			t.Errorf("Expected %v got %v", hex.EncodeToString(hashed), a)
			break
		}
	}
}
func TestStringToHex(t *testing.T) {
	str_id := "0123456789abcdef0123456789abcdef01234567"
	if kademlia.NewID(str_id).String() != str_id {
		t.Error(kademlia.NewID(str_id).String())
	}
}
