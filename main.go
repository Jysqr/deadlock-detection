package main

import (
	"context"
	"deadlock-detection/DeadlockNode"
	"deadlock-detection/MessageTypes"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"time"
)

var (
	numNode     = 3
	bossNode, _ = noise.NewNode(noise.WithNodeBindPort(39999))
	network     = kademlia.New()
)

func main() {
	bossNode.Bind(network.Protocol())
	bossNode.RegisterMessage(MessageTypes.BossToNode{}, MessageTypes.UnmarshalBossToNode)
	bossNode.RegisterMessage(MessageTypes.NodeToBoss{}, MessageTypes.UnmarshalNodeToBoss)
	bossNode.RegisterMessage(MessageTypes.Probe{}, MessageTypes.UnmarshalProbe)
	bossNode.RegisterMessage(MessageTypes.DeadLock{}, MessageTypes.UnmarshalDeadLock)
	bossNode.Handle(func(ctx noise.HandlerContext) error {
		msgObj, err := ctx.DecodeMessage()
		if err != nil {
			panic(err)
		}
		switch msg := msgObj.(type) {
		case MessageTypes.Probe:
		case MessageTypes.NodeToBoss:
			go messageHandler(msg)
		case MessageTypes.BossToNode:
		case MessageTypes.DeadLock:
			panic("DEADLOCK DETECTED") //todo: do thing with deadlock
		}
		return nil
	})
	if err := bossNode.Listen(); err != nil {
		panic(err)
	}

	//get int from gui
	for i := 0; i < numNode; i++ { //builds the nodes and gives them their init parameters
		dn := DeadlockNode.NewDeadlockNode(bossNode.Addr(), numNode)
		go func(node *DeadlockNode.DeadlockNode) {
			dn.Start()
		}(dn)
	}

	time.Sleep(2 * time.Second)

	messageAllCommand("shutdown", "")
	messageAllCommand("step", "")
	messageAllCommand("work", "1")
	time.Sleep(5 * time.Second)
	fmt.Println("exiting..")
}
func messageHandler(msg MessageTypes.NodeToBoss) {
	fmt.Println("report: " + msg.Report)
}

func messageAllCommand(c string, p string) {
	table := network.Table()
	entries := table.Entries()
	for _, con := range entries {
		if err := bossNode.SendMessage(context.TODO(), con.Address, MessageTypes.BossToNode{Command: c, Param: p}); err != nil {
			fmt.Println(err)
		}
	}
}

/*
Possible boss to node Commands:
shutdown
step
work + param
setLocalDependence
*/
