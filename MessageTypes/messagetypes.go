package MessageTypes

import (
	"encoding/json"
	"fmt"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/perlin-network/noise"
)

type MessageInterface interface{}

type BossToNode struct {
	Command string
	Param   string
}

func (e BossToNode) String() string {
	return fmt.Sprintf("command: %s", e.Command)
}

//serializable interface
func (e BossToNode) Marshal() []byte {
	content, _ := json.Marshal(e)
	return content
}

func UnmarshalBossToNode(b []byte) (BossToNode, error) {
	var msg BossToNode
	err := json.Unmarshal(b, &msg)
	return msg, err
}

type NodeToBoss struct {
	Report string
}

func (e NodeToBoss) Marshal() []byte {
	content, _ := json.Marshal(e)
	return content
}

//serializable interface
func UnmarshalNodeToBoss(b []byte) (NodeToBoss, error) {
	var msg NodeToBoss
	err := json.Unmarshal(b, &msg)
	return msg, err
}
func (e NodeToBoss) String() string {
	return fmt.Sprintf("command: %s", e.Report)
}

//probe for deadlock detection algorithm

type Probe struct {
	ProcessI noise.ID
	ProcessJ noise.ID
	ProcessK noise.ID
}

func (e Probe) String() string {
	return fmt.Sprintf("i: %s\nj: %s\n k: %s", e.ProcessI, e.ProcessJ, e.ProcessK)
}

//serializable interface
func (e Probe) Marshal() []byte {
	content, _ := json.Marshal(e)
	return content
}
func UnmarshalProbe(b []byte) (Probe, error) {
	var msg Probe
	err := json.Unmarshal(b, &msg)
	return msg, err
}

type DeadLock struct {
	Deadlock string
}

//serializable interface
func (e DeadLock) Marshal() []byte {
	content, _ := json.Marshal(e)
	return content
}
func UnmarshalDeadLock(b []byte) (DeadLock, error) {
	var msg DeadLock
	err := json.Unmarshal(b, &msg)
	return msg, err
}

type MessageWrapper struct {
	Message MessageInterface
	Sender  noise.ID
}

func (mw MessageWrapper) Compare(other queue.Item) int {
	//noinspection GoImpossibleTypeAssertion
	otherMW := other.(MessageWrapper)
	msg := otherMW.Message
	var returnVal int
	switch v := msg.(type) {
	case Probe:
		returnVal = 1
	case BossToNode:
		returnVal = 0
	case DeadLock:
		returnVal = 1
	case NodeToBoss:
		returnVal = 0
	default:
		returnVal = 0
		v = v //arbitrary assign to clear unused variable error. can't use _ in a switch for some dumb reason
	}
	return returnVal
}
