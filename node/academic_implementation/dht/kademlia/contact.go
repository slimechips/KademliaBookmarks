package kademlia

import "fmt"

type Contact struct {
	Id      ID
	Address string
}

func (contact *Contact) String() string {
	return fmt.Sprintf("Contact(\"%s\", \"%s\")", contact.Id, contact.Address)
}
