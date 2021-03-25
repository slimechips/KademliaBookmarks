package node

import (
	"container/list"
)

type RoutingTable struct {
	Buckets [ID_LENGTH * 8]*list.List
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
	return ret
}

func copyToList(bucket *list.List, list *[]nodeCoreRecord, target ID) {
	for e := bucket.Front(); e != nil; e = e.Next() {
		NodeCore := e.Value.(*NodeCore)
		*list = append(*list, nodeCoreRecord{NodeCore, NodeCore.GUID.Xor(target)})
	}
}
