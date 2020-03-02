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
	network  *kademlia.Protocol
	node     *noise.Node
	err      error
	bossAddr string
	step     *sync.Mutex
}

func NewDeadlockNode(initStruct *InitStruct, s *sync.Mutex) *DeadlockNode {
	initStruct.mutex.Lock() //a listen cannot be called concurrently, so block before listen
	newNode, error := noise.NewNode()
	dn := &DeadlockNode{
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
	dn.node.RegisterMessage(&MessageTypes.BossToNode{}, MessageTypes.UnmarshalBossToNode) //todo figure out while this doesnt seem to do anything
	fmt.Println(dn.node.RegisterMessage(&MessageTypes.NodeToBoss{}, MessageTypes.UnmarshalNodeToBoss))
	return dn
}

func (n *DeadlockNode) Start() {
	for i := 0; i < 6; i++ {
		s := fmt.Sprintf("%d step for %s complete", i, n.node.ID())
		msg := MessageTypes.NodeToBoss{Report: s}
		n.step.Lock()
		fmt.Printf("%s has stepped\n", n.node.ID())
		err := n.node.SendMessage(context.TODO(), n.bossAddr, &msg)
		if err != nil {
			fmt.Println(err)
		}
		n.step.Unlock()

	}

}
