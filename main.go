package main

import (
	"deadlock-detection/ui"
)

var (
	node *ui.GUINode
)

func main() {
	node = ui.NewGUINode(4999)
	node.Run()
}
