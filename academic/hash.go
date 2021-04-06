package academic

import (
	"crypto/sha1"
	"fmt"
)

func NewSHA256(data []byte) []byte {
	hash := sha1.Sum(data)
	return hash[:]
}
func hashFileName(file_name string) []byte {
	file_name_byte_array := []byte(file_name)
	hashed_file := NewSHA256(file_name_byte_array)
	return hashed_file
}
func main() {
	var file_name string = "kademlia_paper"
	hashed_file := hashFileName(file_name)
	fmt.Printf("hashed_file %v", hashed_file)
}
