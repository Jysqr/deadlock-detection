package main

import (
	"deadlock-detection/ui"
	"fmt"
)

var (
	guiNode *ui.GUINode
)

func main() {
	guiNode = ui.NewGUINode(4999)
	fmt.Println(guiNode.BossNode.Addr())
	guiNode.Run()

}
