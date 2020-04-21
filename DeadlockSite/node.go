package DeadlockSite

import (
	"context"
	"deadlock-detection/MessageTypes"
	"fmt"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"math"
	"math/rand"
	"strconv"
	"time"
)

type DeadlockNode struct {
	online              bool
	block               bool
	wait                bool
	workToken           int
	totalNodeCount      int
	network             *kademlia.Protocol
	node                *noise.Node
	bossAddr            string
	dependant           map[string]bool //string is a node address
	selfLocalDependence bool
	messageQueue        *queue.PriorityQueue
	workQueue           *queue.Queue
	siteDependency      noise.ID       //this is for the algorithm
	dependsOnMe         []noise.ID     //this is to know where to send completed work.
	workReqList         map[string]int //string is a node address, fills with work completed by other nodes. producer/consumer
	MySite              *Site
}

func NewDeadlockNode(address string, count int, site *Site, depend noise.ID) *DeadlockNode {
	newNode, _ := noise.NewNode()
	dn := &DeadlockNode{
		online:              true,
		wait:                true,
		totalNodeCount:      count,
		block:               false,
		selfLocalDependence: false,
		network:             kademlia.New(),
		node:                newNode,
		bossAddr:            address,
		workToken:           -1,
		siteDependency:      depend,
		MySite:              site,
		messageQueue:        queue.NewPriorityQueue(50),
		workReqList:         make(map[string]int),
		dependant:           make(map[string]bool),
		dependsOnMe:         make([]noise.ID, site.TotalNodeCount),
		workQueue:           queue.New(50),
	}
	dn.node.Bind(dn.network.Protocol())

	dn.node.RegisterMessage(MessageTypes.BossToNode{}, MessageTypes.UnmarshalBossToNode)
	dn.node.RegisterMessage(MessageTypes.NodeToBoss{}, MessageTypes.UnmarshalNodeToBoss)
	dn.node.RegisterMessage(MessageTypes.Probe{}, MessageTypes.UnmarshalProbe)
	dn.node.RegisterMessage(MessageTypes.DeadLock{}, MessageTypes.UnmarshalDeadLock)

	dn.node.Handle(func(ctx noise.HandlerContext) error { //anonymous function that determines what happens when a message is received
		msgObj, err := ctx.DecodeMessage()
		senderID := ctx.ID()
		if err != nil {
			panic(err)
		}
		message, ok := msgObj.(MessageTypes.MessageInterface) //checks if the message isn't malformed
		if ok {
			wrapper := MessageTypes.MessageWrapper{
				Message: message,
				Sender:  senderID,
			}
			_ = dn.messageQueue.Put(wrapper)
		}
		return nil
	})
	dn.workReqList[depend.Address] = 0
	if err := dn.node.Listen(); err != nil {
		panic(err)
	}
	if _, err := dn.node.Ping(context.TODO(), dn.bossAddr); err != nil { //making sure that node can communicate with boss
		panic(err) //this error is fatal for the node.
	}
	return dn
}

func (dn *DeadlockNode) Start() {
	for dn.online {
		go dn.emptyMessageQueue() //goroutine to parse messages into work and tokens
		if dn.messageQueue.Empty() && dn.workQueue.Empty() {
			s := fmt.Sprintf("node %s is waiting", dn.node.ID().Address)
			dn.messageBoss(s)
			dn.wait = true
		} else if !dn.workQueue.Empty() && dn.workToken != 0 {
			dn.workToken = dn.workToken - 1
			value, err := dn.workQueue.Get(1)
			if err != nil {
				fmt.Println(err)
			} else {
				dn.doWork(value[0].(int))
			}
			s := fmt.Sprintf("node %s is waiting after completing work", dn.node.ID().Address)
			dn.messageBoss(s)
			dn.wait = true
		} else if dn.workQueue.Empty() || dn.workToken == 0 {

		}

	}
	for dn.wait {
		//waiting for next step command
	}

	fmt.Println("Node Shut Down")
}

func (dn *DeadlockNode) doWork(work int) {
	//randomly chose prime number function as the work function
	//intentionally unoptimized
	var prime = false
	for i := 2; i <= int(math.Floor(float64(work)/2)); i++ {
		if work%i == 0 {
			prime = false
		}
	}
	if prime { //if the value is prime, send a token to all dependants
		for _, dependant := range dn.dependsOnMe {
			dn.messageNode(dependant.Address, "produced")
		}
	} else { //if its not prime, send one token to a random dependant
		rand.Seed(time.Now().UnixNano())
		randomPos := rand.Intn(len(dn.dependsOnMe) - 1)
		dependant := dn.dependsOnMe[randomPos]
		dn.messageNode(dependant.Address, "produced")
	}
}

func (dn *DeadlockNode) emptyMessageQueue() {
	if !dn.messageQueue.Empty() {
		items, err := dn.messageQueue.Get(dn.messageQueue.Len() - 1)
		if err != nil {
			fmt.Println(err)
		} else {
			for _, item := range items {
				messageWrap := item.(MessageTypes.MessageWrapper)
				message := messageWrap.Message
				sender := messageWrap.Sender
				switch v := message.(type) {
				case MessageTypes.DeadLock:
					//todo deadlock shit
				case MessageTypes.Probe:
					dn.receiveProbe(v)
				case MessageTypes.BossToNode:
					dn.respondCommand(v.Command, v.Param, sender)
				default:
					//intentional empty statement
				}
			}
		}
	}
}
func (dn *DeadlockNode) respondCommand(command string, p string, sender noise.ID) {
	switch command {
	case "shutdown":
		s := fmt.Sprintf("node %s is shutting down", dn.node.ID().Address)
		dn.messageBoss(s)
		dn.online = false
	case "step":
		s := fmt.Sprintf("node %s is commencing a step", dn.node.ID().Address)
		dn.messageBoss(s)
		dn.wait = false
	case "setup":
		s := fmt.Sprintf("node %s is commencing setup", dn.node.ID().Address)
		dn.messageBoss(s)
		dn.setupOutSiteDepend()
		if len(dn.workReqList) != 0 {
			dn.workToken = 0 //enabling worktokens
		}
	case "work":
		if len(dn.workReqList) > 0 {
			var ready = true
			for _, value := range dn.workReqList { //check if all the required work has been complete
				if value == 0 {
					ready = false
				}
			}
			if ready {
				for key, value := range dn.workReqList { //reduce all counts by 1 and create a token
					dn.workReqList[key] = value - 1
				}
				dn.workToken = dn.workToken + 1
			}
			value, err := strconv.Atoi(p)
			if err != nil {
				fmt.Println(err)
			} else {
				_ = dn.workQueue.Put(value)
			}
		}
	case "produced":
		value := dn.workReqList[sender.Address]
		dn.workReqList[sender.Address] = value + 1
	case "setLocalDependence":
		s := fmt.Sprintf("node %s has enabled Self Local Dependence", dn.node.ID().Address)
		dn.messageBoss(s)
		dn.selfLocalDependence = true
	case "depend":
		dn.workReqList[sender.Address] = 0
		dn.workToken = 0 //enabling worktokens
	}
}

func (dn *DeadlockNode) setupOutSiteDepend() {
	//this function randomly decides a node that will depend on THIS node
	dn.network.Discover() //scan for all other nodes
	table := dn.network.Table()
	entries := table.Entries()       //collect all node ID
	rand.Seed(time.Now().UnixNano()) //generate seed using current time
	var randomDependant noise.ID
	var foundDependant bool
	var outsideSite bool
	for !foundDependant {
		randomPos := rand.Intn(len(entries) - 1)
		randomDependant = entries[randomPos]
		outsideSite = true
		for _, mySiteNode := range dn.MySite.NodeList { //checks if the node picked is within this site or is this node
			if mySiteNode.node.ID().Address == randomDependant.Address {
				outsideSite = false
			}
		}
		if randomDependant.Address != dn.bossAddr && outsideSite { //checking if its the boss
			foundDependant = true
		}
	}
	dn.dependsOnMe = append(dn.dependsOnMe, randomDependant)
	dn.messageNode(randomDependant.Address, "depend")
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
	if !dn.block && !dn.dependant[probe.ProcessI.String()] && !dn.messageQueue.Empty() {
		dn.dependant[probe.ProcessI.String()] = true
		myID := dn.node.ID()
		if myID.String() == probe.ProcessI.String() {
			dn.messageAllDeadlock("Cyclical wait detected")
		} else {
			dn.sendProbe(probe.ProcessI, probe.ProcessI, probe.ProcessI) //todo this is completely wrong
		}
	}
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

func (dn *DeadlockNode) messageBoss(report string) {
	if err := dn.node.SendMessage(context.TODO(), dn.bossAddr, MessageTypes.NodeToBoss{Report: report}); err != nil {
		panic(err)
	}
}
func (dn *DeadlockNode) messageNode(address string, message string) {
	if err := dn.node.SendMessage(context.TODO(), address,
		MessageTypes.BossToNode{
			Command: message,
			Param:   "",
		}); err != nil {
		fmt.Println(err)
	}
}