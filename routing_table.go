package main

import (
	"container/list"
	"sync"
)

/*
A routing table contains List<LinkedList<NodeCore>>, information regarding other nodes in the network.
The mutex lock ensures multiple requests do not modify routing table a the same time
*/
type RoutingTable struct {
	Buckets [ID_LENGTH * 8]*list.List // List of K-Buckets in the routing table. Each K-Bucket contains linked list of node cores
	Mutex   *sync.Mutex               // Mutex lock for routing table
}

type nodeCoreRecord struct {
	Node    *NodeCore
	SortKey ID
}

func NewRoutingTable() (ret *RoutingTable) {
	ret = new(RoutingTable)
	for i := 0; i < ID_LENGTH*8; i++ {
		ret.Buckets[i] = list.New()
	}
	ret.Mutex = &sync.Mutex{}
	return
}

func addToList(node *NodeCore, list *[]nodeCoreRecord, target ID) {
	*list = append(*list, nodeCoreRecord{node, node.GUID.Xor(target)})
}

func copyToList(bucket *list.List, list *[]nodeCoreRecord, target ID) {
	for e := bucket.Front(); e != nil; e = e.Next() {
		NodeCore := e.Value.(*NodeCore)
		*list = append(*list, nodeCoreRecord{NodeCore, NodeCore.GUID.Xor(target)})
	}
}
