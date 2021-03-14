// routing table

package kademlia

import (
	"container/list"
	"sort"
)

const BucketSize = 20

type RoutingTable struct {
	Node    Contact
	Buckets [IdLength * 8]*list.List
}

type ContactRecord struct {
	Node    *Contact
	SortKey ID
}

func NewRoutingTable(node *Contact) (ret *RoutingTable) {
	ret = new(RoutingTable)
	for i := 0; i < IdLength*8; i++ {
		ret.Buckets[i] = list.New()
	}
	ret.Node = *node
	return
}

func (table *RoutingTable) Update(contact *Contact) {
	prefix_length := contact.Id.Xor(table.Node.Id).PrefixLen()

	bucket := table.Buckets[prefix_length]

	var element *list.Element = nil
	for e := bucket.Front(); e != nil; e = e.Next() {
		if e.Value.(*Contact).Id.Equals(contact.Id) {
			element = e
			break
		}
	}
	if element == nil {
		if bucket.Len() <= BucketSize {
			bucket.PushFront(contact)
		}
		// TODO: Handle insertion when the list is full by evicting old elements if
		// they don't respond to a ping.
	} else {
		bucket.MoveToFront(element)
	}
}

func copyToList(bucket *list.List, list *[]ContactRecord, target ID) {
	for e := bucket.Front(); e != nil; e = e.Next() {
		contact := e.Value.(*Contact)
		*list = append(*list, ContactRecord{contact, contact.Id.Xor(target)})
	}
}

func (table *RoutingTable) FindClosest(target ID, count int) (ret []ContactRecord) {
	bucket_num := target.Xor(table.Node.Id).PrefixLen()
	bucket := table.Buckets[bucket_num]
	copyToList(bucket, &ret, target)

	// bidirectional search from the middle node
	for i := 1; (bucket_num-i >= 0 || bucket_num+i < IdLength*8) && len(ret) < count; i++ {
		if bucket_num-i >= 0 {
			bucket = table.Buckets[bucket_num-i]
			copyToList(bucket, &ret, target)
		}
		if bucket_num+1 < IdLength*8 {
			bucket = table.Buckets[bucket_num+i]
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

// 	element := iterable.Find(bucket, func(x interface{}) bool {
// 		return x.(*Contact).id.Equals(table.node.id)
// 	})

// 	if element == nil {
// 		if bucket.Len() <= BucketSize {
// 			bucket.PushFront(contact)
// 		}
// 		// TODO: Handle insertion when the list is full by evicting old elements if
// 		// they don't respond to a ping.
// 	} else {
// 		bucket.MoveToFront(element.(*list.Element))
// 	}
// }

// func copyToVector(start, end *list.Element, vec *vector.Vector, target ID) {
// 	for elt := start; elt != end; elt = elt.Next() {
// 		contact := elt.Value.(*Contact)
// 		vec.Push(&ContactRecord{contact, contact.id.Xor(target)})
// 	}
// }

// func (table *RoutingTable) FindClosest(target ID, count int) (ret *vector.Vector) {
// 	ret = new(vector.Vector).Resize(0, count)

// 	bucket_num := target.Xor(table.node.id).PrefixLen()
// 	bucket := table.buckets[bucket_num]
// 	copyToVector(bucket.Front(), nil, ret, target)

// 	for i := 1; (bucket_num-i >= 0 || bucket_num+i < IdLength*8) && ret.Len() < count; i++ {
// 		if bucket_num-i >= 0 {
// 			bucket = table.buckets[bucket_num-i]
// 			copyToVector(bucket.Front(), nil, ret, target)
// 		}
// 		if bucket_num+i < IdLength*8 {
// 			bucket = table.buckets[bucket_num+i]
// 			copyToVector(bucket.Front(), nil, ret, target)
// 		}
// 	}

// 	sort.Sort(ret)
// 	if ret.Len() > count {
// 		ret.Cut(count, ret.Len())
// 	}
// 	return
// }
