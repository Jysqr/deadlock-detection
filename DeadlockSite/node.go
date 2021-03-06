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

//represents a node in a network
type DeadlockNode struct {
	online              bool
	block               bool
	wait                bool
	workToken           int
	totalNodeCount      int
	network             *kademlia.Protocol
	node                *noise.Node
	bossAddr            string
	dependant           map[string]bool //string is a node address, this is for the algorithm
	selfLocalDependence bool
	messageQueue        *queue.PriorityQueue
	workQueue           *queue.Queue
	dependsOnMe         []string       //this is to know where to send completed work.
	producerList        map[string]int //string is a node address, fills with work completed by other nodes. producer/consumer
	workWaitList        []string
	MySite              *Site
	MyAddress           string
}

func NewDeadlockNode(address string, count int, site *Site, siteDepend string) *DeadlockNode {
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
		MySite:              site,
		messageQueue:        queue.NewPriorityQueue(50),
		producerList:        make(map[string]int),
		dependant:           make(map[string]bool),
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
	if siteDepend != "" {
		dn.dependsOnMe[0] = siteDepend
		dn.messageNode(siteDepend, "depend")
	}

	return dn
}

func (dn *DeadlockNode) Start() {
	if err := dn.node.Listen(); err != nil {
		panic(err)
	}
	s := fmt.Sprintf("node %s has started", dn.node.ID().Address)
	dn.messageBoss(s, 0)
	dn.MyAddress = dn.node.Addr()
	stop := make(chan struct{})
	dn.block = true
	go dn.emptyMessageQueue(stop) //goroutine to parse messages into work and tokens
	for dn.block {                //wait for setup command to come and be completed

	}
	for dn.online {
		for dn.wait {
			//waiting for next step command
		}
		if dn.selfLocalDependence {
			dn.sendProbe(dn.node.ID().Address) //node has no tokens, time to check for deadlock
		} else if dn.messageQueue.Empty() && dn.workQueue.Empty() {
			s := fmt.Sprintf("node %s is waiting due to no work", dn.node.ID().Address)
			dn.messageBoss(s, 1)
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
			dn.messageBoss(s, 2)
			dn.wait = true
		} else if !dn.workQueue.Empty() && dn.workToken == 0 {
			for key, value := range dn.producerList {
				if value == 0 {
					dn.workWaitList = append(dn.workWaitList, key)
				}
			}
			dn.sendProbe(dn.node.ID().Address) //node has no tokens, time to check for deadlock
			s := fmt.Sprintf("node %s is waiting after unable to complete work (no token)", dn.node.ID().Address)
			dn.messageBoss(s, 1)

			dn.wait = true
		}
	}
	close(stop)
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
	var sentTo string
	if prime { //if the value is prime, send a token to all dependants
		for _, dependant := range dn.dependsOnMe {
			dn.messageNode(dependant, "produced")
			sentTo = sentTo + dependant + " "
		}
	} else { //if its not prime, send one token to a random dependant
		rand.Seed(time.Now().UnixNano())
		randomPos := rand.Intn(len(dn.dependsOnMe))
		dependant := dn.dependsOnMe[randomPos]
		dn.messageNode(dependant, "produced")
	}

	s := fmt.Sprintf("%s", sentTo)
	dn.messageBoss(s, -2)
}

func (dn *DeadlockNode) emptyMessageQueue(done <-chan struct{}) {
	for { //for loops keeps the goroutine running
		select {
		case <-done: //if the channel closes, the goroutine is shutdown
			return
		default: //this is always ran unless the channel is closed
			if !dn.messageQueue.Empty() {
				items, err := dn.messageQueue.Get(dn.messageQueue.Len())
				if err != nil {
					fmt.Println(err)
				} else {
					for _, item := range items {
						messageWrap := item.(MessageTypes.MessageWrapper)
						message := messageWrap.Message
						sender := messageWrap.Sender
						switch v := message.(type) {
						case MessageTypes.DeadLock:
							dn.online = false
							dn.messageBoss("deadlock declared", 4)
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
	}
}

func (dn *DeadlockNode) respondCommand(command string, p string, sender noise.ID) {
	switch command {
	case "shutdown":
		s := fmt.Sprintf("node %s is shutting down", dn.node.ID().Address)
		dn.messageBoss(s, 0)
		dn.online = false
	case "step":
		dn.wait = false
	case "setup":
		dn.setupOutSiteDepend()
		s := fmt.Sprintf("node %s begins setup", dn.node.ID().Address)
		dn.messageBoss(s, 0)
		if len(dn.producerList) != 0 {
			dn.workToken = 0 //enabling worktokens
		}
		dn.block = false
		s = fmt.Sprintf("node %s complete3 setup", dn.node.ID().Address)
		dn.messageBoss(s, 1)
	case "work":
		if len(dn.producerList) > 0 {
			var ready = true
			for _, value := range dn.producerList { //check if all the required work has been complete
				if value == 0 {
					ready = false
				}
			}
			if ready {
				for key, value := range dn.producerList { //reduce all counts by 1 and create a token
					dn.producerList[key] = value - 1
				}
				dn.workToken = dn.workToken + 1
			}
		}
		value, err := strconv.Atoi(p)
		if err != nil {
			fmt.Println(err)
		} else {
			_ = dn.workQueue.Put(value)
		}
	case "produced":
		value := dn.producerList[sender.Address]
		dn.producerList[sender.Address] = value + 1
		var tempWorkWaitList []string
		copy(tempWorkWaitList, dn.workWaitList)
		counter := 0
		if tempWorkWaitList != nil {
			for i, address := range dn.workWaitList {
				if address == sender.Address {
					counter = counter + 1
					tempWorkWaitList[len(tempWorkWaitList)-1], tempWorkWaitList[i] = tempWorkWaitList[i], tempWorkWaitList[len(tempWorkWaitList)-1]
				}
			}
		}
		dn.workWaitList = tempWorkWaitList[:len(tempWorkWaitList)-counter]
	case "setLocalDependence":
		s := fmt.Sprintf("node %s has enabled Self Local Dependence", dn.node.ID().Address)
		dn.messageBoss(s, -1)
		dn.selfLocalDependence = true
	case "depend":
		dn.producerList[sender.Address] = 0
		dn.workToken = 0 //enabling worktokens
		s := fmt.Sprintf("node %s added a producer %s", dn.node.ID().Address, sender.Address)
		dn.messageBoss(s, 1)
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
	dn.dependsOnMe = append(dn.dependsOnMe, randomDependant.Address)
	dn.messageNode(randomDependant.Address, "depend")
}

func (dn *DeadlockNode) sendProbe(processI string) {
	/*
	   	Process of sending probe:

	      1. If process Pi is locally dependent on itself then declare a deadlock.

	      2. Else for all Pj and Pk check following condition:

	          (a). Process Pi is locally dependent on process Pj
	          (b). Process Pj is waiting on process Pk

	      If all of the above conditions are true, send probe (i, j, k) to the home site of process Pk.
	*/
	if dn.selfLocalDependence {
		s := fmt.Sprintf("%s is locally dependent on itself", dn.node.ID().Address)
		dn.messageAllDeadlock(s)
	} else {
		processKList := make(map[string]string)
		for key := range dn.producerList { //workreqlist is all the j that this node are locally dependent on
			for _, site := range dn.MySite.SiteList { //searching every site
				for _, node := range site.NodeList { //searching every node within each site
					if node.node.ID().Address == key { //4 nested for each loops. nice.
						if node.wait {
							for _, address := range node.workWaitList {
								processKList[key] = address //key is the processk, value is the processJ it connects to
							}
						}
					}
				}
			}
		}
		if len(processKList) > 0 {
			for processK, processJ := range processKList {
				probe := MessageTypes.Probe{
					ProcessI: processI,
					ProcessJ: processJ,
					ProcessK: processK,
				}
				if err := dn.node.SendMessage(context.TODO(), probe.ProcessK, probe); err != nil {
					panic(err)
				}
				s := fmt.Sprintf("node %s sent probe to %s", dn.node.ID().Address, probe.ProcessK)
				dn.messageBoss(s, 3)
				s = fmt.Sprintf("%s", probe.ProcessK)
				dn.messageBoss(s, -3)
			}

		}
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

	      4. Send probe (i, m, n) to the home site of process Pn if above conditions satisfy.

	*/
	//process k above is this node
	if !dn.block && !dn.dependant[probe.ProcessI] && !dn.workQueue.Empty() {
		dn.dependant[probe.ProcessI] = true
		myID := dn.node.ID()
		if myID.Address == probe.ProcessI {
			dn.messageAllDeadlock("Cyclical wait detected")
		} else {
			processKList := make(map[string]string)
			for key := range dn.producerList { //workreqlist is all the j that this node are locally dependent on
				for _, site := range dn.MySite.SiteList { //searching every site
					for _, node := range site.NodeList { //searching every node within each site
						if node.node.ID().Address == key { //4 nested for each loops. nice.
							if node.wait {
								for _, address := range node.workWaitList {
									processKList[key] = address //key is the processk, value is the processJ it connects to
								}
							}
						}
					}
				}
			}
			if len(processKList) > 0 {
				for processK, processJ := range processKList {
					newProbe := MessageTypes.Probe{
						ProcessI: probe.ProcessI,
						ProcessJ: processJ,
						ProcessK: processK,
					}
					if err := dn.node.SendMessage(context.TODO(), newProbe.ProcessK, probe); err != nil {
						panic(err)
					}
					s := fmt.Sprintf("node %s sent probe after receiving to %s", dn.node.ID().Address, newProbe.ProcessK)
					dn.messageBoss(s, 3)
					s = fmt.Sprintf("%s", probe.ProcessK)
					dn.messageBoss(s, -3)
				}

			}
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

func (dn *DeadlockNode) messageBoss(report string, status int) {
	if err := dn.node.SendMessage(context.TODO(), dn.bossAddr, MessageTypes.NodeToBoss{Report: report, Status: status}); err != nil {
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
