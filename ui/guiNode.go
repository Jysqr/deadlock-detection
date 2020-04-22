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
	"math/rand"
	"os"
	"strconv"
	"time"
)

var ( //tunables for timings
	stepMinTime            = 5     //in seconds
	workSendInterval       = 5     //equals (stepMinTime * workSendInterval) seconds
	maxPrimeCandidateValue = 25000 //these two are for the work sent to the nodes
	minPrimeCandidateValue = 500
)

type GUINode struct {
	frontend       *DeadlockDetectionSimulator
	app            *widgets.QApplication
	numSite        int
	numNodePerSite int
	BossNode       *noise.Node
	network        *kademlia.Protocol
	siteList       []*DeadlockSite.Site
	totalNode      int
	step           bool
	stop           chan struct{}
	//junk
	x int
	y int
}

func NewGUINode(port uint16) *GUINode {
	newNode, err := noise.NewNode()
	if err != nil {
		fmt.Println(err)
	}
	guiNode := &GUINode{
		app:      widgets.NewQApplication(len(os.Args), os.Args),
		frontend: NewDeadlockDetectionSimulator(nil),
		BossNode: newNode,
		network:  kademlia.New(),
		step:     false,
		stop:     make(chan struct{}),
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
	if err := guiNode.BossNode.Listen(); err != nil {
		panic(err)
	}
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
		if err := gNode.BossNode.SendMessage(context.TODO(), con.Address, MessageTypes.BossToNode{Command: c, Param: p}); err != nil {
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
		gNode.stop = make(chan struct{})
		for i := 0; i < gNode.numSite; i++ { //builds the nodes and gives them their init parameters
			newSite := DeadlockSite.NewSite(gNode.BossNode.Addr(), gNode.numNodePerSite, gNode.totalNode, gNode.siteList[:])
			gNode.siteList[i] = newSite
		}
		if gNode.step {
			gNode.frontend.StartPauseButton.SetText("Step")
		} else {
			gNode.frontend.StartPauseButton.SetText("Pause")
			go gNode.autoStep(gNode.stop)
		}
	case "Pause":
		gNode.frontend.StartPauseButton.SetText("Start")
		close(gNode.stop)
	case "Step":
		gNode.frontend.StepCheckBox.SetDisabled(true)
		gNode.frontend.StartPauseButton.SetDisabled(true)
		gNode.messageAllCommand("step", "")
		time.Sleep(time.Duration(stepMinTime) * time.Second)
		gNode.frontend.StepCheckBox.SetDisabled(false)
		gNode.frontend.StartPauseButton.SetDisabled(false)
	}
}

func (gNode *GUINode) autoStep(stop <-chan struct{}) {
	counter := 0
	for {
		select {
		default:
			gNode.messageAllCommand("step", "")
			if counter%workSendInterval == 0 {
				rand.Seed(time.Now().UnixNano())
				randPrimeCandidate := rand.Intn(maxPrimeCandidateValue-minPrimeCandidateValue+1) + minPrimeCandidateValue
				gNode.messageAllCommand("work", strconv.Itoa(randPrimeCandidate))
			}
			counter = counter + 1
		case <-stop: // triggered when the stop channel is closed
			break // exit
		}
		time.Sleep(time.Duration(stepMinTime) * time.Second)
	}

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
