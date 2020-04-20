package MessageTypes

import (
	"encoding/json"
	"fmt"
	"github.com/perlin-network/noise"
)

type BossToNode struct {
	Command string
	Param   string
}

func (e BossToNode) Marshal() []byte {
	content, _ := json.Marshal(e)
	return content
}

func (e BossToNode) String() string {
	return fmt.Sprintf("command: %s", e.Command)
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

func UnmarshalNodeToBoss(b []byte) (NodeToBoss, error) {
	var msg NodeToBoss
	err := json.Unmarshal(b, &msg)
	return msg, err
}
func (e NodeToBoss) String() string {
	return fmt.Sprintf("command: %s", e.Report)
}

type Probe struct {
	ProcessI noise.ID
	ProcessJ noise.ID
	ProcessK noise.ID
}

func (e Probe) String() string {
	return fmt.Sprintf("i: %s\nj: %s\n k: %s", e.ProcessI, e.ProcessJ, e.ProcessK)
}

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

func (e DeadLock) Marshal() []byte {
	content, _ := json.Marshal(e)
	return content
}
func UnmarshalDeadLock(b []byte) (DeadLock, error) {
	var msg DeadLock
	err := json.Unmarshal(b, &msg)
	return msg, err
}
