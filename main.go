package main

import (
	"deadlock-detection/ui"
)

//starts everything
var (
	guiNode *ui.GUINode
)

func main() {
	guiNode = ui.NewGUINode(4999)
	guiNode.Run()

}
