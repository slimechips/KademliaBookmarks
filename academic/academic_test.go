package main

import (
	"fmt"
	"main/academic/node"
	"testing"
)

func TestItem(t *testing.T) {
	item := node.NewItem("New")
	fmt.Print(item.Value)
	fmt.Println("RANDOM")
}

func main() {
	item := node.NewItem("New")
	fmt.Print(item.Value)
	fmt.Println("RANDOM")
}
