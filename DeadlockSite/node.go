package DeadlockSite

import (
	"context"
	"deadlock-detection/MessageTypes"
	"deadlock-detection/resource"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
)

type DeadlockNode struct {
	online              bool
	block               bool
	wait                bool
	totalNodeCount      int
	network             *kademlia.Protocol
	node                *noise.Node
	err                 error
	bossAddr            string
	dependant           map[string]bool
	selfLocalDependence bool
	queue               []noise.Serializable
	currentWaitList     []noise.ID
	siteDependency      noise.ID
	MySite              *Site
}

func NewDeadlockNode(address string, count int, site *Site, depend noise.ID) *DeadlockNode {
	newNode, noiseError := noise.NewNode()
	dn := &DeadlockNode{
		online:              true,
		wait:                true,
		totalNodeCount:      count,
		block:               false,
		selfLocalDependence: false,
		network:             kademlia.New(),
		node:                newNode,
		err:                 noiseError,
		bossAddr:            address,
		siteDependency:      depend,
		MySite:              site,
	}
	dn.node.Bind(dn.network.Protocol())

	dn.node.RegisterMessage(MessageTypes.BossToNode{}, MessageTypes.UnmarshalBossToNode)
	dn.node.RegisterMessage(MessageTypes.NodeToBoss{}, MessageTypes.UnmarshalNodeToBoss)
	dn.node.RegisterMessage(MessageTypes.Probe{}, MessageTypes.UnmarshalProbe)
	dn.node.RegisterMessage(MessageTypes.DeadLock{}, MessageTypes.UnmarshalDeadLock)

	dn.node.Handle(func(ctx noise.HandlerContext) error {
		msgObj, err := ctx.DecodeMessage()
		if err != nil {
			panic(err)
		}
		deadlock, ok := msgObj.(MessageTypes.DeadLock) //checks if the message is alertign deadlock, takes precedence
		if ok {
			panic(deadlock.Deadlock + "DEADLOCK DETECTED") //todo do somehing with deadlock
		} else {
			dn.queue = dn.enqueue(dn.queue, msgObj)
		}
		return nil
	})
	if err := dn.node.Listen(); err != nil {
		panic(err)
	}
	if _, err := dn.node.Ping(context.TODO(), dn.bossAddr); err != nil {
		panic(err)
	}
	return dn
}

func (dn *DeadlockNode) Start() {

	for dn.online {
		if len(dn.queue) == 0 {
			s := fmt.Sprintf("node %s is waiting", dn.node.ID().Address)
			if err := dn.node.SendMessage(context.TODO(), dn.bossAddr, MessageTypes.NodeToBoss{Report: s}); err != nil {
				panic(err)
			}
			dn.wait = true
		} else {
			for len(dn.queue) != 0 { //do all work sent
				fmt.Println("queuetime")
				var msgObj noise.Serializable
				dn.queue, msgObj = dn.dequeue(dn.queue)
				switch v := msgObj.(type) {
				case MessageTypes.Probe:
					dn.receiveProbe(v)
				case MessageTypes.BossToNode:
					dn.respondCommand(v.Command, v.Param)
				}
			}
		}
		for dn.wait {
			//waiting for next continue command
		}

	}
	fmt.Println("Node Shut Down")
}

func (dn *DeadlockNode) respondCommand(c string, p string) {
	switch c {
	case "shutdown":
		s := fmt.Sprintf("node %s is shutting down", dn.node.ID().Address)
		if err := dn.node.SendMessage(context.TODO(), dn.bossAddr, MessageTypes.NodeToBoss{Report: s}); err != nil {
			panic(err)
		}
		dn.online = false
	case "step":
		dn.wait = false
	case "work":
		switch p {
		case "1":
			if resource.LockResourceOne() {
				dn.enqueue(dn.queue, MessageTypes.BossToNode{
					Command: "release",
					Param:   "1",
				})
			} else {

			}
		case "2":
			if resource.LockResourceTwo() {
				dn.enqueue(dn.queue, MessageTypes.BossToNode{
					Command: "release",
					Param:   "2",
				})
			}
		}
	case "setLocalDependence":
		dn.selfLocalDependence = true
	case "release":
		switch p {
		case "1":
			resource.UnlockResourceOne()
		case "2":
			resource.UnlockResourceTwo()
		}
	}
}

func (dn *DeadlockNode) sendProbe(i noise.ID, j noise.ID, k noise.ID) {
	/*
	   	Process of sending probe:

	      1. If process Pi is locally dependent on itself then declare a deadlock.

	      2. Else for all Pj and Pk check following condition:

	          (a). Process Pi is locally dependent on process Pj
	          (b). Process Pj is waiting on process Pk
	          (c). Process Pj and process Pk are on different sites.

	      If all of the above conditions are true, send probe (i, j, k) to the home site of process Pk.
	*/
	if dn.selfLocalDependence {
		s := fmt.Sprintf("%s is locally dependent", dn.node.ID().Address)
		dn.messageAllDeadlock(s)
	} else {
		//grouping sites probably required
		//create probe after checking waitlist n shit
	}
}

func (dn *DeadlockNode) receiveProbe(probe MessageTypes.Probe) {

	/*
	   	Process Pk checks the following conditions:


	          (a). Process Pk is blocked.
	          (b). dependentk[i] is false.
	          (c). Process Pk has not replied to all requests of process Pj

	      If all of the above conditions are found to be true then:

	      1. Set dependentk[i] to true.
	      2. Now, If k == i then, declare the Pi is deadlocked.
	      3. Else for all Pm and Pn check following conditions:

	          (a). Process Pk is locally dependent on process Pm and
	          (b). Process Pm is waiting upon process Pn and
	          (c). Process Pm and process Pn are on different sites.

	      4. Send probe (i, m, n) to the home site of process Pn if above conditions satisfy.

	*/
	if !dn.block && !dn.dependant[probe.ProcessI.String()] && len(dn.queue) != 0 {
		dn.dependant[probe.ProcessI.String()] = true
		myID := dn.node.ID()
		if myID.String() == probe.ProcessI.String() {
			dn.messageAllDeadlock("Cyclical wait detected")
		} else {
			dn.sendProbe(probe.ProcessI, probe.ProcessI, probe.ProcessI) //todo this is completely wrong
		}
	}
}

func (dn *DeadlockNode) enqueue(queue []noise.Serializable, element noise.Serializable) []noise.Serializable {
	queue = append(queue, element) // append to enqueue.
	fmt.Println("adding to queue")
	return queue
}

func (dn *DeadlockNode) dequeue(queue []noise.Serializable) ([]noise.Serializable, noise.Serializable) { //slices are weird
	element := queue[0]       // The first element is the one to be dequeued.
	return queue[1:], element // Slice off the element once it is dequeued.
}

func (dn *DeadlockNode) messageAllDeadlock(reason string) {
	//this func declares deadlock to all nodes it has seen
	table := dn.network.Table() //grabs full network table
	entries := table.Entries()  //pulls list of connections from the table
	//iterate through list of connections, discarding position
	for _, con := range entries {
		//send message to address, if the error exists print the error
		if err := dn.node.SendMessage(context.TODO(), con.Address,
			MessageTypes.DeadLock{Deadlock: reason}); err != nil {
			fmt.Println(err)
		}
	}
}
