package academic

import (
	"academic/node"
	"fmt"
)

func main() {
	item1 := node.NewItem("Normal")
	fmt.Println(item1)

	item2 := node.NewItemALT("Alternate")
	n1 := node.NewNode(1)
	fmt.Println(n1)

	n1.Data[1] = item1
	n1.Data[2] = item2
	n1.Listen()
	for {

	}
}
