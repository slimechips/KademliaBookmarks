package kademlia
import (
	"testing"
)
///////////////////////
/*helper functions*////
///////////////////////
func CompareOnlyElementInArrayIfEqual(arr1 []nodeCore, arr2 []nodeCore)bool{
	if(len(arr1)==len(arr2) && len(arr1)==1 && arr1[0]==arr2[0]){
		return true
	}else{
		return false
	}
}

func FindNodeCoreInOtherNode(node_core nodeCore, n node) bool {
	prefix_length := node_core.GUID.Xor(n.node_core.GUID).PrefixLen()

	bucket := n.routing_table.Buckets[prefix_length]
	for e := bucket.Front(); e != nil; e = e.Next() {
		if e.Value.(*nodeCore).GUID.Equals((node_core).GUID) {
			return true
		}
	}
	return false
}


/////////////////////////
///////////////////////
/*node generator functions*////
///////////////////////

func TestCreateNodeCore(t *testing.T){
    guid := NewRandomID()
	node0 := createNodeCore(guid,172,1)
	if node0.GUID != guid {
		t.Errorf( "createNodeCore(guid,172,1) FAILED, expected %v but got value %v", node0.GUID, guid )
	} else {
        
		t.Logf( "createNodeCore(guid,172,1) PASSED, expected %v and got value %v", node0.GUID, guid )
	}
}

func TestCreateNode(t *testing.T){
    guid := NewRandomID()
	node0 := createNode(guid,172,1,true)
    if node0.node_core.GUID != guid {
		t.Errorf( "createNode(guid,172,1) FAILED, expected %v but got value %v", node0.node_core.GUID, guid )
	}
}

func TestBroadcast(t *testing.T){
	guid_0 := NewRandomID()
	guid_1 := NewRandomID()
	node0 := createNode(guid_0,172,1,true)
	node1 := createNode(guid_1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	const k int = 1
	bucket:=node0.FindClosest(guid_1,k)
	if bucket[0] != node1.node_core {
		t.Errorf( "joinBroadcastNode(joiningNode *node, broadcastNode *node) FAILED, expected %v but got value %v",  bucket[0], node1.node_core )
	}
}

func TestReqBroadcast(t *testing.T){
	guid_0 := NewRandomID()
	guid_1 := NewRandomID()
	node0 := createNode(guid_0,172,1,true)
	node1 := createNode(guid_1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	// node2 := createNode(2,172,1,true)
	node1 = reqBroadcastNode(node1,node0)

	prefix_length:=node0.node_core.GUID.Xor(node1.node_core.GUID).PrefixLen()
	if !FindNodeCoreInOtherNode(node0.node_core, *node1) {
		t.Errorf( "reqBroadcastNode(joiningNode *node, broadcastNode *node)) FAILED, expected %v but got value %v", node0.node_core,  node1.routing_table.Buckets[prefix_length].Front().Value)
	}
	// bucket_node0:=node0.FindClosest(guid_1,k)
}

// 	if !(CompareOnlyElementInArrayIfEqual(bucket_node1,node0.routing_table)){
// 		t.Errorf( "reqBroadcastNode(joiningNode *node, broadcastNode *node)) FAILED, expected %v but got value %v", node2.routing_table.bucket0, node0.routing_table.bucket0 )
// 	}
// }

// ///////////////////////////
// /*test RPC for kademlia*///
// ///////////////////////////
func TestStore(t *testing.T){
	key := NewRandomID()
	value := "wikipedia"
	guid_0 := NewRandomID()
	node0 := createNode(guid_0,172,1,true)
	node0 = store(node0,key,value)
	if node0.data[key] != value{
		t.Errorf( "store(node0,key,value) FAILED, expected %v but got value %v", node0.data[key], key)
	}
}
func TestPing(t *testing.T){
	const alive_state bool = false
	guid_0 := NewRandomID()
	node0 := createNode(guid_0,172,1,alive_state)
	status := ping(node0)
	if !(node0.alive == alive_state && status ==alive_state){
		t.Errorf( "ping(receipientNode *node) FAILED, expected %v but got value %v", node0.alive , alive_state)
	}
}

func TestFindNode(t *testing.T){
	guid_0 := NewRandomID()
	guid_1 := NewRandomID()
	node0 := createNode(guid_0,172,1,true)
	node1 := createNode(guid_1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	nodeCoreList := findNode(node0,node1.node_core.GUID)
	if !nodeCoreList[0].GUID.Equals(node1.node_core.GUID) {
		t.Errorf( "findNode(receipientNode *node, key ID) FAILED, expected %v but got value %v", node1.node_core.GUID , nodeCoreList[0].GUID)
	}
}
func TestFindValuesInNode(t *testing.T){
	guid_0 := NewRandomID()
	guid_1 := NewRandomID()
	key := guid_1
	const value string = "SCREAMMMMM"
	node0 := createNode(guid_0,172,1,true)
	node1 := createNode(guid_1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	node0 = store(node0,key,value)
	data,flag, k_bucket_list := findValue(node0,key)
	if !(flag==true && data == node0.data[key] && k_bucket_list==nil){
		t.Errorf( "findValue(receipientNode *node,key int)) FAILED, expected %v but got value %v", data , node0.data[key])
	}
}
func TestFindValuesNotInNode(t *testing.T){
	guid_0 := NewRandomID()
	guid_1 := NewRandomID()
	key := guid_1
	node0 := createNode(guid_0,172,1,true)
	node1 := createNode(guid_1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	_ ,flag, k_bucket_list := findValue(node0,key)

	nodeCoreList := findNode(node0,node1.node_core.GUID)
	if !(k_bucket_list[0].GUID.Equals(node1.node_core.GUID) && flag ==false) {
		t.Errorf( "findNode(receipientNode *node, key ID) FAILED, expected %v but got value %v", node1.node_core.GUID , nodeCoreList[0].GUID)
	}
}
