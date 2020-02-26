package MessageTypes

import "encoding/json"

type BossToNode struct {
	Command string
}

func (e *BossToNode) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func UnmarshalBinaryBossToNode(b []byte) *BossToNode {
	var pBossToNode *BossToNode
	_ = json.Unmarshal(b, pBossToNode)
	return pBossToNode
}

type NodeToBoss struct {
	Report string
}

func (e *NodeToBoss) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func UnmarshalBinaryNodeToBoss(b []byte) *NodeToBoss {
	var pNodeToBoss *NodeToBoss
	_ = json.Unmarshal(b, pNodeToBoss)
	return pNodeToBoss
}
