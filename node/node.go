package node

import (
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
)

var (
	nodeNum   int
	network   = kademlia.New()
	node, err = noise.NewNode()
)

func start(num int) {
	if err != nil {
		panic(err)
	}
	node.Bind(network.Protocol())

}
