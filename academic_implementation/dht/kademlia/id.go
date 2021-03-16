package kademlia

import (
	"crypto/sha1"
	"encoding/hex"
	"math/rand"
)

const IdLength = 20

type ID [IdLength]byte

func NewID(data string) (ret ID) {
	decoded, _ := hex.DecodeString(data)
	for i := 0; i < IdLength; i++ {
		ret[i] = decoded[i]
	}
	return
}

func NewRandomID() (ret ID) {
	for i := 0; i < IdLength; i++ {
		ret[i] = uint8(rand.Intn(256))
	}
	return
}

func NewSHA1ID(address string) (ret ID) {
	h := sha1.New()
	h.Write([]byte(address))
	hashed := h.Sum(nil)
	for i := 0; i < IdLength; i++ {
		ret[i] = hashed[i]
	}
	return
}
func (id ID) String() string {
	return hex.EncodeToString(id[0:IdLength])
}

func (id ID) Equals(other ID) bool {
	for i := 0; i < IdLength; i++ {
		if id[i] != other[i] {
			return false
		}
	}
	return true
}

func (id ID) Xor(other ID) (ret ID) {
	for i := 0; i < IdLength; i++ {
		ret[i] = id[i] ^ other[i]
	}
	return
}

func (id ID) PrefixLen() (ret int) {
	for i := 0; i < IdLength; i++ {
		for j := 0; j < 8; j++ {
			if (id[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}
	return IdLength*8 - 1
}

func (id ID) Less(other interface{}) bool {
	for i := 0; i < IdLength; i++ {
		if id[i] != other.(ID)[i] {
			return id[i] < other.(ID)[i]
		}
	}
	return false
}
