package main

import (
	"fmt"
	"net"
)

/*
Basic information describing a node
*/
type NodeCore struct {
	GUID ID
	IP   net.IP
	Port int
}

/*
Return String representation of a NodeCore which can be sent through UDP Network
*/
func (node *NodeCore) String() string {
	return fmt.Sprintf("%s~%s~%d~;", node.GUID.String(), node.IP.String(), node.Port)
}
