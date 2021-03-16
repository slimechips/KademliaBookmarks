package kademlia

import "fmt"

type Contact struct {
	Id      ID
	Address string
}

func NewContact(address string) *Contact {
	id := NewSHA1ID(address)
	return &Contact{Id: id, Address: address}
}
func (contact *Contact) String() string {
	return fmt.Sprintf("Contact(\"%s\", \"%s\")", contact.Id, contact.Address)
}
