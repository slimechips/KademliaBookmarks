// routing table

package kademlia

import (
	"container/list"
	"sort"
)

const BucketSize = 20

type RoutingTable struct {
	Buckets [IdLength * 8]*list.List
}

type nodeCoreRecord struct {
	Node    *nodeCore
	SortKey ID
}

func NewRoutingTable() (ret *RoutingTable) {
	ret = new(RoutingTable)
	for i := 0; i < IdLength*8; i++ {
		ret.Buckets[i] = list.New()
	}
	return
}

func (n *node) Update(otherNodeCore *nodeCore) {
	prefix_length := otherNodeCore.GUID.Xor(n.node_core.GUID).PrefixLen()

	bucket := n.routing_table.Buckets[prefix_length]

	var element *list.Element = nil
	for e := bucket.Front(); e != nil; e = e.Next() {
		if e.Value.(nodeCore).GUID.Equals((otherNodeCore).GUID) {
			element = e
			break
		}
	}
	if element == nil {
		if bucket.Len() <= BucketSize {
			bucket.PushFront(otherNodeCore)
		}
		// TODO: Handle insertion when the list is full by evicting old elements if
		// they don't respond to a ping.

	} else {
		bucket.MoveToFront(element)
	}
}

func copyToList(bucket *list.List, list *[]nodeCoreRecord, target ID) {
	for e := bucket.Front(); e != nil; e = e.Next() {
		nodeCore := e.Value.(*nodeCore)
		*list = append(*list, nodeCoreRecord{nodeCore, nodeCore.GUID.Xor(target)})
	}
}

func (n *node) FindClosest(target ID, count int) (ret []nodeCore){
	records := n.FindClosestRecord(target, count)
	for i := range records {
		ret = append(ret, *records[i].Node)
	}
	return
}

func (n *node) FindClosestRecord(target ID, count int) (ret []nodeCoreRecord) {
	bucket_num := target.Xor(n.node_core.GUID).PrefixLen()
	bucket := n.routing_table.Buckets[bucket_num]
	copyToList(bucket, &ret, target)

	// bidirectional search from the middle node
	for i := 1; (bucket_num-i >= 0 || bucket_num+i < IdLength*8) && len(ret) < count; i++ {
		if bucket_num-i >= 0 {
			bucket = n.routing_table.Buckets[bucket_num-i]
			copyToList(bucket, &ret, target)
		}
		if bucket_num+1 < IdLength*8 {
			bucket = n.routing_table.Buckets[bucket_num+i]
			copyToList(bucket, &ret, target)
		}
	}

	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].SortKey.Less(ret[j].SortKey)
	})

	// slice if somehow the list is longer than what is needed
	if len(ret) > count {
		ret = ret[:count]
	}

	return ret
}

