package main

import (
	"context"
	"deadlock-detection/DeadlockSite"
	"deadlock-detection/MessageTypes"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"time"
)

type GUINode struct {
	numSite        int
	numNodePerSite int
	bossNode       *noise.Node
	network        *kademlia.Protocol
	siteList       []*DeadlockSite.Site
	totalNode      int
}

func NewGUINode(site int, nodes int) *GUINode {
	newNode, _ := noise.NewNode(noise.WithNodeBindPort(39999))
	guiNode := &GUINode{
		numSite:        site,
		numNodePerSite: nodes,
		bossNode:       newNode,
		network:        kademlia.New(),
		totalNode:      site * nodes,
	}
	return guiNode
}

func (gNode *GUINode) main() {
	gNode.bossNode.Bind(gNode.network.Protocol())
	gNode.bossNode.RegisterMessage(MessageTypes.BossToNode{}, MessageTypes.UnmarshalBossToNode)
	gNode.bossNode.RegisterMessage(MessageTypes.NodeToBoss{}, MessageTypes.UnmarshalNodeToBoss)
	gNode.bossNode.RegisterMessage(MessageTypes.Probe{}, MessageTypes.UnmarshalProbe)
	gNode.bossNode.RegisterMessage(MessageTypes.DeadLock{}, MessageTypes.UnmarshalDeadLock)
	gNode.bossNode.Handle(func(ctx noise.HandlerContext) error {
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
	if err := gNode.bossNode.Listen(); err != nil {
		panic(err)
	}

	for i := 0; i < gNode.numSite; i++ { //builds the nodes and gives them their init parameters
		newSite := DeadlockSite.NewSite(gNode.bossNode.Addr(), gNode.numNodePerSite, gNode.totalNode)
		gNode.siteList = append(gNode.siteList, newSite)
	}

	time.Sleep(2 * time.Second)

	gNode.messageAllCommand("shutdown", "")
	gNode.messageAllCommand("step", "")
	gNode.messageAllCommand("work", "1")
	time.Sleep(5 * time.Second)
	fmt.Println("exiting..")
}
func messageHandler(msg MessageTypes.NodeToBoss) {
	fmt.Println("report: " + msg.Report)
}

func (gNode *GUINode) messageAllCommand(c string, p string) {
	table := gNode.network.Table()
	entries := table.Entries()
	for _, con := range entries {
		if err := gNode.bossNode.SendMessage(context.TODO(), con.Address, MessageTypes.BossToNode{Command: c, Param: p}); err != nil {
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
