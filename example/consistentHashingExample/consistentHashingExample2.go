package main

import (
	"fmt"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/node"
)

func main() {
	//create ring using NewRing
	ring := ch.NewRing()

	//Add some nodes to consistent hashing ring 
	ring.AddSync(node.NewNode("alice","127.0.0.1", 0, 3))
	ring.AddSync(node.NewNode("bob","127.0.0.1", 1, 3))
	ring.AddSync(node.NewNode("charlie","127.0.0.1", 2, 3))
	ring.AddSync(node.NewNode("daphnie","127.0.0.1", 3, 3))
	ring.AddSync(node.NewNode("eric","127.0.0.1", 4, 3))

	//faka request
	photoData := "abcdefg1234567"

	//show all keys
	fmt.Printf("All keys:\n%v\n\n", ring.Keys())

	//Get key
	photoKey := ring.Hash(photoData)
	fmt.Printf("Hash key of request: \n%s\n\n", photoKey)

	//Current key
	n, k, e := ring.LookUp(photoKey)
	fmt.Printf("current node is %+v, key is %+v, err is %+v\n\n", n, k, e)

	//Successor
	n, k, e = ring.Successor(k)
	fmt.Printf("successor node is %+v, key is %+v, err is %+v\n\n", n, k, e)

	//2nd Successor
	n, k, e = ring.Successor(k)
	fmt.Printf("2nd successor node is %+v, key is %+v, err is %+v\n\n", n, k, e)

}
