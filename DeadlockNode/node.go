package DeadlockNode

import (
	"context"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"sync"
)

type InitStruct struct {
	mutex   sync.Mutex
	Address string
}
type DeadlockNode struct {
	network  *kademlia.Protocol
	node     *noise.Node
	err      error
	bossAddr string
}

func NewDeadlockNode(initStruct *InitStruct) *DeadlockNode {
	initStruct.mutex.Lock() //a listen cannot be called concurrently, so this is the only time
	newNode, error := noise.NewNode()
	dn := &DeadlockNode{
		network:  kademlia.New(),
		node:     newNode,
		err:      error,
		bossAddr: initStruct.Address,
	}
	dn.node.Bind(dn.network.Protocol())
	initStruct.mutex.Unlock()
	if err := dn.node.Listen(); err != nil {
		panic(err)
	}
	if _, err := dn.node.Ping(context.TODO(), n.bossAddr); err != nil {
		panic(err)
	}
	return dn
}

func (n *DeadlockNode) Start() {

	n.node.Send(context.TODO(), n.bossAddr, []byte("message sent"))
	fmt.Printf("node %s discovered %d peer(s).\n", n.node.ID().Address, len(n.network.Discover()))
}

func (s *InitStruct) initialize() {

}
