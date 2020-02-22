package node

import (
	"context"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"strconv"
)

var (
	nodeNum int
	network        = kademlia.New()
	port    uint16 = 40000
	node    *noise.Node
	err     error
)

func Start(num int, bossAddr string) {
	nodeNum = num
	fmt.Println(strconv.Itoa(nodeNum) + " " + strconv.Itoa(int(port)+nodeNum))
	node, err = noise.NewNode(noise.WithNodeBindPort(port + uint16(nodeNum)))
	if err != nil {
		panic(err)
	}
	node.Bind(network.Protocol())
	if err := node.Listen(); err != nil {
		panic(err)
	}
	if _, err := node.Ping(context.TODO(), bossAddr); err != nil {
		panic(err)
	}
}
