package main
import "testing"
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
/////////////////////////
///////////////////////
/*node generator functions*////
///////////////////////

func TestCreateNodeCore(t *testing.T){
	const guid int = 0
	node0 := createNodeCore(guid,172,1)
	if node0.GUID != guid {
		t.Errorf( "createNodeCore(guid,172,1) FAILED, expected %v but got value %v", node0.GUID, guid )
	} else {
        
		t.Logf( "createNodeCore(guid,172,1) PASSED, expected %v and got value %v", node0.GUID, guid )
	}
}

func TestCreateNode(t *testing.T){
    const guid int = 0
	node0 := createNode(guid,172,1,true)
    if node0.node_core.GUID != guid {
		t.Errorf( "createNode(guid,172,1) FAILED, expected %v but got value %v", node0.node_core.GUID, guid )
	}
}

func TestBroadcast(t *testing.T){
	node0 := createNode(0,172,1,true)
	node1 := createNode(1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	if node0.routing_table.bucket0[0] != node1.node_core {
		t.Errorf( "joinBroadcastNode(joiningNode *node, broadcastNode *node) FAILED, expected %v but got value %v", node1.node_core, node0.routing_table.bucket0[0] )
	}
}

func TestReqBroadcast(t *testing.T){
	node0 := createNode(0,172,1,true)
	node1 := createNode(1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	node2 := createNode(2,172,1,true)
	node2 = reqBroadcastNode(node2,node0)
	if !(node2.routing_table.bucket0[0] ==node0.routing_table.bucket0[0] && len(node2.routing_table.bucket0)==len(node0.routing_table.bucket0)){
		t.Errorf( "reqBroadcastNode(joiningNode *node, broadcastNode *node)) FAILED, expected %v but got value %v", node2.routing_table.bucket0, node0.routing_table.bucket0 )
	}
}

///////////////////////////
/*test RPC for kademlia*///
///////////////////////////
func TestStore(t *testing.T){
	const key int = 10
	const value int = 100 
	node0 := createNode(0,172,1,true)
	node0 = store(node0,key,value)
	if node0.data[key] != value{
		t.Errorf( "store(node0,key,value) FAILED, expected %v but got value %v", node0.data[key], key)
	}
}
func TestPing(t *testing.T){
	const alive_state bool = false
	node0 := createNode(0,172,1,alive_state)
	status := ping(node0)
	if !(node0.alive == alive_state && status ==alive_state){
		t.Errorf( "ping(receipientNode *node) FAILED, expected %v but got value %v", node0.alive , alive_state)
	}
}
func TestFindNode(t *testing.T){
	node0 := createNode(0,172,1,true)
	node1 := createNode(1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	nodeBucket := findNode(node0,1)
	if !(nodeBucket[0] == node0.routing_table.bucket0[0] && len(nodeBucket)==len(node0.routing_table.bucket0) && len(node0.routing_table.bucket0)==1){
		t.Errorf( "findNode(node0,1) FAILED, expected %v but got value %v", nodeBucket , node0.routing_table.bucket0)
	}
}
func TestFindValuesInNode(t *testing.T){
	const key int = 10
	const value int = 100 
	node0 := createNode(0,172,1,true)
	node1 := createNode(1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	node0 = store(node0,key,value)
	data,flag, k_bucket_list := findValue(node0,key)
	if !(flag==true && data == node0.data[key] && k_bucket_list==nil){
		t.Errorf( "findValue(receipientNode *node,key int)) FAILED, expected %v but got value %v", data , node0.data[key])
	}
}
func TestFindValuesNotInNode(t *testing.T){
	const key int = 10
	node0 := createNode(0,172,1,true)
	node1 := createNode(1,172,1,true)
	node0 = joinBroadcastNode(node1,node0)
	data,flag, k_bucket_list := findValue(node0,key)
	if !(flag==false && CompareOnlyElementInArrayIfEqual(k_bucket_list,node0.routing_table.bucket0)==true ){
		t.Errorf( "findValue(receipientNode *node,key int)) FAILED, expected %v but got value %v", data , node0.data[key])
	}
}
