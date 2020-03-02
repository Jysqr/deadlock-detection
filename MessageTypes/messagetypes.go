package MessageTypes

import (
	"encoding/json"
	"github.com/perlin-network/noise"
)

type BossToNode struct {
	Command string
}

func (e BossToNode) Marshal() []byte {
	content, _ := json.Marshal(e)
	return content
}

func UnmarshalBossToNode(b []byte) (BossToNode, error) {
	var msg BossToNode
	err := json.Unmarshal(b, msg)
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

type Probe struct {
	processI noise.ID
	processJ noise.ID
	processK noise.ID
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
