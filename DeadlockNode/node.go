package DeadlockNode

import (
	"context"
	"deadlockdetection/MessageTypes"
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
	online    bool
	network   *kademlia.Protocol
	node      *noise.Node
	err       error
	bossAddr  string
	step      *sync.Mutex
	dependant []bool
}

func NewDeadlockNode(initStruct *InitStruct, s *sync.Mutex) *DeadlockNode {
	initStruct.mutex.Lock() //a listen cannot be called concurrently, so block before listen
	newNode, error := noise.NewNode()
	dn := &DeadlockNode{
		online:   true,
		network:  kademlia.New(),
		node:     newNode,
		err:      error,
		bossAddr: initStruct.Address,
		step:     s,
	}
	dn.node.Bind(dn.network.Protocol())
	initStruct.mutex.Unlock()
	if err := dn.node.Listen(); err != nil {
		panic(err)
	}
	if _, err := dn.node.Ping(context.TODO(), dn.bossAddr); err != nil {
		panic(err)
	}
	dn.node.RegisterMessage(MessageTypes.BossToNode{}, MessageTypes.UnmarshalBossToNode)
	dn.node.RegisterMessage(MessageTypes.NodeToBoss{}, MessageTypes.UnmarshalNodeToBoss)
	return dn
}

func (dn *DeadlockNode) Start() {
	for i := 0; i < 6; i++ {
		s := fmt.Sprintf("%d step for %s complete", i, dn.node.ID())
		dn.step.Lock()

		if err := dn.node.SendMessage(context.TODO(), dn.bossAddr, MessageTypes.NodeToBoss{Report: s}); err != nil {
			panic(err)
		}
		dn.step.Unlock()

	}

}
