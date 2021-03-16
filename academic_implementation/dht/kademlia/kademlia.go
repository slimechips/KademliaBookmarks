package kademlia
import "fmt"

const (
	k = 5
)
type nodeCore struct {
	GUID ID
	ip_address int
	udp_port int
}
// type routingTable struct {
// 	bucket0 []nodeCore
// 	bucket1 []nodeCore
// }

type node struct{
	node_core nodeCore
	routing_table RoutingTable
	data map[ID] string	//stores a <key,value> pair for retrieval
	alive bool	//for unit testing during prototyping
}

func createNodeCore(guid ID, ip_address int, udp_port int) *nodeCore {
	n := nodeCore{GUID:guid,ip_address:ip_address,udp_port:udp_port}
	return &n
}

func createNode(guid ID, ip_address int, udp_port int, alive bool) *node {
	var node0 *nodeCore = createNodeCore(guid, ip_address, udp_port)
	// var rt0 routingTable = routingTable{bucket0:nil,bucket1:nil}
	var routing_table *RoutingTable = NewRoutingTable()
	return &node{node_core:*node0,routing_table:*routing_table,alive:alive}
}

func joinBroadcastNode(joiningNode *node, broadcastNode *node) *node {
	broadcastNode.Update(&joiningNode.node_core)
	return broadcastNode
}

func reqBroadcastNode(joiningNode *node, broadcastNode *node) *node {
	joiningNode.Update(&broadcastNode.node_core)
	for i:= range broadcastNode.routing_table.Buckets {
		bucket := broadcastNode.routing_table.Buckets[i]
		for e := bucket.Front(); e != nil; e = e.Next() {
			joiningNode.Update(e.Value.(*nodeCore))
		}
	}
	return joiningNode
}

//////////////////////////////////////////
/*RPC for Kademlia *//////////////////////
//////////////////////////////////////////
func store(receipientNode *node, key ID, value string) *node {
	if receipientNode.data == nil{
		receipientNode.data =  make(map[ID]string)
	}
	receipientNode.data[key]=value
	fmt.Println(receipientNode.data[key])
	return receipientNode
}
func ping(receipientNode *node) bool {
	return receipientNode.alive
}
func findNode(receipientNode *node, key ID)[]nodeCore{
	result := receipientNode.FindClosest(key, k)
	var out []nodeCore
	for i := range result {
		out = append(out, *result[i].Node)
	}
	return out
}
func findValue(receipientNode *node,key ID)(string, bool, []nodeCore) {
	if _, ok := receipientNode.data[key]; ok {
		return receipientNode.data[key],true,nil
	}else{
		return "",false,findNode(receipientNode,key)
	}
}




func main(){
	fmt.Println("hi")
	// var node0 nodeCore = nodeCore{GUID:0,ip_address:172,udp_port:8080}
	// var node0 *nodeCore = createNodeCore(0,172,1)
	// fmt.Println(node0.GUID)
	// var rt0 routingTable = routingTable{bucket0:nil,bucket1:nil}
	// rt0.bucket0 = append(rt0.bucket0,node0)
	// fmt.Println(rt0.bucket0)
	// var node1 nodeCore = nodeCore{GUID:1,ip_address:172,udp_port:8079}
	// fmt.Println(node1.GUID)
	// rt0.bucket0 = append(rt0.bucket0,node1)
	// node0 := createNode(0,172,1)
	// fmt.Println(node0.node_core.GUID)
	// node0 := createNode(0,172,1)
	// node1 := createNode(1,172,1)
	// node0 = joinBroadcastNode(node0,node1)
	// fmt.Println(node0.routing_table.bucket0)
	// fmt.Println("node0 bucket core",node0.routing_table.bucket0[0])
	// print("node1 ",node1.node_core)
	// const key int = 10
	// const value int = 100 
	// node0 := createNode(0,172,1)
	// node0 = store(node0,key,value)
	// fmt.Println(node0.data[10])
	
	// node_1
}
