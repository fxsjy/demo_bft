package main

import (
	"demo_bft/node"
	"fmt"
	"time"
)

const blockSpan = 10000 * time.Microsecond //10 ms
func main() {
	network := []*node.Node{node.NewNode("A"), node.NewNode("B"), node.NewNode("C"), node.NewNode("D")}
	for _, node := range network {
		node.SetNeighbours(network)
	}
	nodeA := network[0]
	for i := 0; i < 1000; i++ {
		blkid := fmt.Sprintf("block_%d", i)
		nodeA.ProposeBlock(&node.Block{Blockid: blkid, Height: i})
		time.Sleep(blockSpan)
	}
	//nodeA.ProposeBlock(&node.Block{Blockid:"foo", Height: 1})
	//nodeA.ProposeBlock(&node.Block{Blockid:"bar", Height: 1})
	time.Sleep(3 * time.Second)
	for _, node := range network {
		fmt.Println(node.Id, len(node.Blocks))
	}
}
