package ui

import (
	"context"
	"deadlock-detection/DeadlockSite"
	"deadlock-detection/MessageTypes"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"os"
	"strconv"
)

type GUINode struct {
	frontend       *DeadlockDetectionSimulator
	app            *widgets.QApplication
	numSite        int
	numNodePerSite int
	bossNode       *noise.Node
	network        *kademlia.Protocol
	siteList       []*DeadlockSite.Site
	totalNode      int
	step           bool
	x              int
	y              int
}

func NewGUINode(port uint16) *GUINode {
	newNode, _ := noise.NewNode(noise.WithNodeBindPort(port))

	guiNode := &GUINode{
		app:      widgets.NewQApplication(len(os.Args), os.Args),
		frontend: NewDeadlockDetectionSimulator(nil),
		bossNode: newNode,
		network:  kademlia.New(),
		step:     false,
	}

	newNode.Bind(guiNode.network.Protocol())
	newNode.RegisterMessage(MessageTypes.BossToNode{}, MessageTypes.UnmarshalBossToNode)
	newNode.RegisterMessage(MessageTypes.NodeToBoss{}, MessageTypes.UnmarshalNodeToBoss)
	newNode.RegisterMessage(MessageTypes.Probe{}, MessageTypes.UnmarshalProbe)
	newNode.RegisterMessage(MessageTypes.DeadLock{}, MessageTypes.UnmarshalDeadLock)
	newNode.Handle(func(ctx noise.HandlerContext) error {
		msgObj, err := ctx.DecodeMessage()
		if err != nil {
			panic(err)
		}
		switch msg := msgObj.(type) {
		case MessageTypes.Probe:
		case MessageTypes.NodeToBoss:
			go messageHandler(msg)
		case MessageTypes.BossToNode:
		case MessageTypes.DeadLock:
			panic("DEADLOCK DETECTED") //todo: do thing with deadlock
		}
		return nil
	})
	guiNode.frontend.OkButton.ConnectClicked(guiNode.okPressed)
	guiNode.frontend.StartPauseButton.ConnectClicked(guiNode.startPressed)
	guiNode.frontend.DrawArea.ConnectPaintEvent(guiNode.linePaintevent)
	guiNode.frontend.DrawArea.ConnectMousePressEvent(guiNode.paint)
	guiNode.frontend.StepCheckBox.ConnectClicked(guiNode.stepChecked)
	guiNode.frontend.StartPauseButton.SetDisabled(true)
	return guiNode
}

func (gNode *GUINode) Run() {
	gNode.frontend.Show()
	gNode.app.Exec()
}

func messageHandler(msg MessageTypes.NodeToBoss) {
	fmt.Println("report: " + msg.Report)
}

func (gNode *GUINode) messageAllCommand(c string, p string) {
	table := gNode.network.Table()
	entries := table.Entries()
	for _, con := range entries {
		if err := gNode.bossNode.SendMessage(context.TODO(), con.Address, MessageTypes.BossToNode{Command: c, Param: p}); err != nil {
			fmt.Println(err)
		}
	}
}

func (gNode *GUINode) okPressed(bool) {
	gNode.frontend.Numberinfo.Hide()
	gNode.numSite = gNode.frontend.SiteSpinBox.Value()
	gNode.numNodePerSite = gNode.frontend.NodeSpinBox.Value()
	gNode.totalNode = gNode.numSite * gNode.numNodePerSite
	gNode.frontend.StartPauseButton.SetDisabled(false)
}

func (gNode *GUINode) startPressed(bool) {
	switch gNode.frontend.StartPauseButton.Text() {
	case "Start":
		for i := 0; i < gNode.numSite; i++ { //builds the nodes and gives them their init parameters
			newSite := DeadlockSite.NewSite(gNode.bossNode.Addr(), gNode.numNodePerSite, gNode.totalNode)
			gNode.siteList = append(gNode.siteList, newSite)
		}
		if gNode.step {
			gNode.frontend.StartPauseButton.SetText("Step")
		} else {
			gNode.frontend.StartPauseButton.SetText("Pause")
		}
	case "Pause":
		gNode.frontend.StartPauseButton.SetText("Start")
		//todo pause work
	case "Step":
		//todo stepstuff

	}
	/*
		time.Sleep(2 * time.Second)
		gNode.messageAllCommand("shutdown", "")
		time.Sleep(5 * time.Second)
		fmt.Println("exiting..")
	*/
}

func (gNode *GUINode) sendWork(target noise.ID) {
	//todo prime number stuff
}

func (gNode *GUINode) stepChecked(bool) {
	gNode.step = gNode.frontend.StepCheckBox.IsChecked()
}

func (gNode *GUINode) paint(event *gui.QMouseEvent) {
	gNode.x = event.X()
	gNode.y = event.Y()
	gNode.frontend.Update()
}

func (gNode *GUINode) linePaintevent(event *gui.QPaintEvent) { /// line numbers
	painter := gui.NewQPainter2(gNode.frontend.DrawArea)
	painter.FillRect3(event.Rect(), gui.NewQBrush3(gui.QColor_FromRgb(90), core.Qt__SolidPattern))
	painter.DrawText3(gNode.x, gNode.y, strconv.Itoa(5))
	painter.DestroyQPainter()
}

func (gNode *GUINode) reset() {
	fmt.Println("reset")

}

/*
Possible boss to node Commands:
shutdown
step
work + param
setLocalDependence
*/
