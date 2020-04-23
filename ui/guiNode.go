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
	"strings"
	"time"
)

//this contains the GUI and bossNode
var ( //tunables for timings
	stepMinTime            = 3     //in seconds
	workSendInterval       = 2     //equals (stepMinTime * workSendInterval) seconds
	maxPrimeCandidateValue = 25000 //these two are for the work sent to the nodes
	minPrimeCandidateValue = 500
)

type GUINode struct {
	frontend         *DeadlockDetectionSimulator
	app              *widgets.QApplication
	numSite          int
	numNodePerSite   int
	BossNode         *noise.Node
	network          *kademlia.Protocol
	siteList         []*DeadlockSite.Site
	totalNode        int
	step             bool
	stop             chan struct{}
	deadlockDeclared bool
	//image stuff
	statusArray         [][]int
	nodeMap             [][]string
	nodePosMap          [][]*core.QPoint
	visibleMessagesList []messageRepresentation
	draw                bool
	firstRun            bool
	resourcePath        string
	nodeImageTop        *gui.QImage
	nodeImageBot        *gui.QImage
	nodeImageLeft       *gui.QImage
	nodeImageRight      *gui.QImage
	drawNodeAddr        bool
}

type messageRepresentation struct {
	start      *core.QPoint
	end        *core.QPoint
	current    *core.QPoint
	iterations int
	color      string
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
		firstRun: true,
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
			go guiNode.messageHandler(msg, ctx.ID().Address)
		case MessageTypes.BossToNode:
		case MessageTypes.DeadLock:
			guiNode.deadlockDeclared = true
			guiNode.frontend.StatusBar.ShowMessage("DEADLOCK", 0)
		}
		return nil
	})
	guiNode.frontend.OkButton.ConnectClicked(guiNode.okPressed)
	guiNode.frontend.StartPauseButton.ConnectClicked(guiNode.startPressed)
	guiNode.frontend.DrawArea.ConnectPaintEvent(guiNode.paintEvent)
	guiNode.frontend.StepCheckBox.ConnectClicked(guiNode.stepChecked)
	guiNode.frontend.StartPauseButton.SetDisabled(true)
	guiNode.frontend.ConsoleCheckBox.SetHidden(true) //turning off unfinished feature
	if err := guiNode.BossNode.Listen(); err != nil {
		panic(err)
	}
	return guiNode
}

func (gNode *GUINode) Run() {
	gNode.frontend.Show()
	gNode.app.Exec()
}

func (gNode *GUINode) messageHandler(msg MessageTypes.NodeToBoss, address string) {
	if msg.Status >= -1 { //if the code is less than this it is a visual message
		fmt.Println("report: " + msg.Report)
		go gNode.updateStatusArray(address, msg.Status) //ran on a goroutine
	} else {
		gNode.visualMessageBuilder(msg.Report, address, msg.Status)
	}
	gNode.frontend.DrawArea.Update()
}

func (gNode *GUINode) visualMessageBuilder(to string, from string, code int) {
	var color string
	var list []messageRepresentation
	var points []*core.QPoint
	splitString := strings.Split(to, " ")
	for _, String := range splitString { //searching for where the receivers are
		for i := 0; i < gNode.numSite; i++ {
			for j := 0; j < gNode.numNodePerSite; j++ {
				if gNode.nodeMap[i][j] == String {
					points = append(points, gNode.nodePosMap[i][j])
				}

			}
		}
	}
	var fromPoint *core.QPoint //the point that is the end position
	for i := 0; i < gNode.numSite; i++ {
		for j := 0; j < gNode.numNodePerSite; j++ {
			if gNode.nodeMap[i][j] == from {
				fromPoint = gNode.nodePosMap[i][j]
			}

		}
	}
	switch code {
	case -2:
		color = "black"
	case -3:
		color = "blue"
	}
	for _, point := range points { //wrapping the points
		temp := messageRepresentation{
			start:      point,
			end:        fromPoint,
			current:    point,
			iterations: 0,
			color:      color,
		}
		list = append(list, temp) //adding to a list for further processing
	}
	if len(list) > 0 {
		gNode.visibleMessagesList = list
	}
}

func (gNode *GUINode) updateStatusArray(addr string, status int) {
	if status > -1 { //finds the node on the list and updates the array
		for i := 0; i < gNode.numSite; i++ {
			for j := 0; j < gNode.numNodePerSite; j++ {
				if gNode.nodeMap[i][j] == addr {
					gNode.statusArray[i][j] = status
				}
			}
		}
	}
}

func (gNode *GUINode) buildNodeMap() {
	for i := 0; i < gNode.numSite; i++ { //puts all the node addresses into a 2d array for easier referencing
		for j := 0; j < gNode.numNodePerSite; j++ {
			gNode.nodeMap[i][j] = gNode.siteList[i].NodeList[j].MyAddress
		}
	}
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
	//gNode.frontend.Numberinfo.Hide()
	gNode.numSite = gNode.frontend.SiteSpinBox.Value()
	gNode.numNodePerSite = gNode.frontend.NodeSpinBox.Value()
	gNode.totalNode = gNode.numSite * gNode.numNodePerSite
	gNode.siteList = make([]*DeadlockSite.Site, gNode.numSite)
	gNode.statusArray = make([][]int, gNode.numSite)
	for j := 0; j < gNode.numSite; j++ {
		gNode.statusArray[j] = make([]int, gNode.numNodePerSite)
	}
	gNode.nodeMap = make([][]string, gNode.numSite)
	for j := 0; j < gNode.numSite; j++ {
		gNode.nodeMap[j] = make([]string, gNode.numNodePerSite)
	}
	gNode.nodePosMap = make([][]*core.QPoint, gNode.numSite)
	for j := 0; j < gNode.numSite; j++ {
		gNode.nodePosMap[j] = make([]*core.QPoint, gNode.numNodePerSite)
	}
	gNode.draw = true
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	gNode.resourcePath = mydir + "/resources"
	switch gNode.numNodePerSite {
	case 2: //i tried to thin this down so it didnt need to be so verbose, but it doesn't work any other way
		//go qt is buggy
		tempPath := gNode.resourcePath + "/node-2b.png"
		gNode.nodeImageBot = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-2t.png"
		gNode.nodeImageTop = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-2r.png"
		gNode.nodeImageRight = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-2l.png"
		gNode.nodeImageLeft = gui.NewQImage9(tempPath, "")
	case 3:
		tempPath := gNode.resourcePath + "/node-3b.png"
		gNode.nodeImageBot = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-3t.png"
		gNode.nodeImageTop = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-3r.png"
		gNode.nodeImageRight = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-3l.png"
		gNode.nodeImageLeft = gui.NewQImage9(tempPath, "")
	case 4:
		tempPath := gNode.resourcePath + "/node-4b.png"
		gNode.nodeImageBot = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-4t.png"
		gNode.nodeImageTop = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-4r.png"
		gNode.nodeImageRight = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-4l.png"
		gNode.nodeImageLeft = gui.NewQImage9(tempPath, "")
	case 5:
		tempPath := gNode.resourcePath + "/node-5b.png"
		gNode.nodeImageBot = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-5t.png"
		gNode.nodeImageTop = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-5r.png"
		gNode.nodeImageRight = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-5l.png"
		gNode.nodeImageLeft = gui.NewQImage9(tempPath, "")
	case 6:
		tempPath := gNode.resourcePath + "/node-6b.png"
		gNode.nodeImageBot = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-6t.png"
		gNode.nodeImageTop = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-6r.png"
		gNode.nodeImageRight = gui.NewQImage9(tempPath, "")
		tempPath = gNode.resourcePath + "/node-6l.png"
		gNode.nodeImageLeft = gui.NewQImage9(tempPath, "")
	}
	gNode.frontend.DrawArea.Update()
	gNode.frontend.StartPauseButton.SetDisabled(false)
}

func (gNode *GUINode) startPressed(bool) {
	switch gNode.frontend.StartPauseButton.Text() {
	case "Start":
		gNode.frontend.StatusBar.ShowMessage("Loading", 0)
		gNode.stop = make(chan struct{})
		for i := 0; i < gNode.numSite; i++ { //builds the nodes and gives them their init parameters
			newSite := DeadlockSite.NewSite(gNode.BossNode.Addr(), gNode.numNodePerSite, gNode.totalNode, gNode.siteList[:])
			gNode.siteList[i] = newSite
		}
		time.Sleep(time.Duration(stepMinTime) * time.Second)
		go gNode.buildNodeMap()
		gNode.messageAllCommand("setup", "")
		gNode.drawNodeAddr = true
		gNode.frontend.DrawArea.Update()
		time.Sleep(time.Duration(stepMinTime) * time.Second)
		if gNode.step {
			gNode.frontend.StartPauseButton.SetText("Step")
		} else {
			gNode.frontend.StartPauseButton.SetText("Pause")
			go gNode.autoStep(gNode.stop)
		}
		gNode.frontend.StatusBar.ShowMessage("Running", 0)
	case "Pause":
		gNode.frontend.StartPauseButton.SetText("Resume")
		close(gNode.stop)
		gNode.frontend.StatusBar.ShowMessage("Paused", 0)
	case "Resume":
		gNode.stop = make(chan struct{})
		go gNode.autoStep(gNode.stop)
		gNode.frontend.StartPauseButton.SetText("Pause")
		gNode.frontend.StatusBar.ShowMessage("Running", 0)
	case "Step":
		gNode.frontend.StatusBar.ShowMessage("Step Mode Enabled", 0)
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
	for { //this function is written to be ran on a goroutine. it sends a message to all nodes every stepMinTime seconds to progress in work
		select {
		default:
			gNode.messageAllCommand("step", "")
			if counter%workSendInterval == 0 {
				fmt.Println("Work has been dispatched")
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

func (gNode *GUINode) paintEvent(event *gui.QPaintEvent) {
	painter := gui.NewQPainter2(gNode.frontend.DrawArea)
	painter.SetPen2(gui.NewQColor6("black"))
	if gNode.draw && !gNode.deadlockDeclared {
		heightDrawArea := gNode.frontend.DrawArea.Height()
		widthDrawArea := gNode.frontend.DrawArea.Width()
		imageHeight := gNode.nodeImageTop.Height()
		imageWidth := gNode.nodeImageTop.Width()
		xCoordMid := widthDrawArea/2 - imageWidth/2
		xCoordLeft := 100
		xCoordRight := widthDrawArea - 100
		rightY := heightDrawArea/2 - imageWidth/2
		botY := heightDrawArea - 100
		topY := 100
		leftY := heightDrawArea / 2
		siteNumOffset := 75
		statusBoxDistance := 50       //the distance between the boxes that get coloured in
		switch gNode.numNodePerSite { //switches dont cascade in go. I hate it.
		case 2:
			painter.DrawImage8(core.NewQPoint2(xCoordMid, botY), gNode.nodeImageBot)
			painter.DrawText3(xCoordMid+42, topY+13, "1")
			painter.DrawImage8(core.NewQPoint2(xCoordMid, topY), gNode.nodeImageTop)
			painter.DrawText3(xCoordMid+42, botY+26, "2")

			if gNode.drawNodeAddr { //to keep magic numbers easier
				//top
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+statusBoxDistance*i, topY+imageHeight+10, gNode.siteList[0].NodeList[i].MyAddress)
				}
				//bottom
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+(statusBoxDistance*i), botY-5, gNode.siteList[1].NodeList[i].MyAddress)
				}
				if gNode.numSite >= 3 {
					//left
					painter.Save()
					painter.Translate3(float64(xCoordLeft), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, -(imageHeight + 5), gNode.siteList[2].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
				//right
				if gNode.numSite == 4 {
					painter.Save()
					painter.Translate3(float64(xCoordRight), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, 12, gNode.siteList[3].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
			}

			switch gNode.numSite {
			case 2:
			case 3:
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, leftY), gNode.nodeImageLeft)
				painter.Save()
				painter.Translate3(float64(xCoordLeft), float64(leftY)) //messing with the coord system is awful
				painter.Rotate(-float64(90))
				painter.DrawText3(-25, 13, "3")
				painter.Restore()
			case 4:
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, leftY), gNode.nodeImageLeft)
				painter.Save()
				painter.Translate3(float64(xCoordLeft), float64(leftY)) //messing with the coord system is awful
				painter.Rotate(-float64(90))
				painter.DrawText3(-25, 13, "3")
				painter.Restore()
				painter.DrawImage8(core.NewQPoint2(xCoordRight, rightY), gNode.nodeImageRight)
				painter.Save()
				painter.Translate3(float64(xCoordRight), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(float64(90))
				painter.DrawText3(siteNumOffset-30, -17, "4")
				painter.Restore()
			}

			gNode.drawStatus(painter, rightY+5, botY, topY+20, leftY+5, xCoordMid+5, xCoordLeft+20, xCoordRight)

		case 3:
			painter.DrawImage8(core.NewQPoint2(xCoordMid, botY), gNode.nodeImageBot) //2 are always drawn
			painter.DrawText3(xCoordMid+siteNumOffset-3, topY+13, "1")
			painter.DrawImage8(core.NewQPoint2(xCoordMid, topY), gNode.nodeImageTop)
			painter.DrawText3(xCoordMid+siteNumOffset-3, botY+26, "2")

			if gNode.drawNodeAddr { //to keep magic numbers easier
				//top
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+(statusBoxDistance*i), topY+imageHeight+10, gNode.siteList[0].NodeList[i].MyAddress)
				}
				//bottom
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+(statusBoxDistance*i), botY-5, gNode.siteList[1].NodeList[i].MyAddress)
				}
				if gNode.numSite >= 3 {
					//left
					painter.Save()
					painter.Translate3(float64(xCoordLeft), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, -(imageHeight + 5), gNode.siteList[2].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
				//right
				if gNode.numSite == 4 {
					painter.Save()
					painter.Translate3(float64(xCoordRight), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, 12, gNode.siteList[3].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
			}
			switch gNode.numSite {
			case 2:
			case 3:
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, rightY), gNode.nodeImageLeft)
				painter.Save()
				painter.Translate3(float64(xCoordLeft), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(float64(90))
				painter.DrawText3(siteNumOffset-4, -7, "3")
				painter.Restore()
			case 4:
				painter.Save()
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, rightY), gNode.nodeImageLeft)
				painter.Translate3(float64(xCoordLeft), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(float64(90))
				painter.DrawText3(siteNumOffset-4, -7, "3")
				painter.Restore()
				painter.DrawImage8(core.NewQPoint2(xCoordRight, rightY), gNode.nodeImageRight)
				painter.Save()
				painter.Translate3(float64(xCoordRight), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(float64(90))
				painter.DrawText3(siteNumOffset-4, -17, "4")
				painter.Restore()
			}

			gNode.drawStatus(painter, rightY, botY, topY, leftY-60, xCoordMid, xCoordLeft, xCoordRight)

		case 4:
			painter.DrawImage8(core.NewQPoint2(xCoordMid, botY), gNode.nodeImageBot) //2 are always drawn
			painter.DrawText3(xCoordMid+siteNumOffset+10, topY+13, "1")
			painter.DrawImage8(core.NewQPoint2(xCoordMid, topY), gNode.nodeImageTop)
			painter.DrawText3(xCoordMid+siteNumOffset+13, botY+26, "2")

			if gNode.drawNodeAddr { //to keep magic numbers easier
				//top
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+(statusBoxDistance*i), topY+imageHeight+10, gNode.siteList[0].NodeList[i].MyAddress)
				}
				//bottom
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+(statusBoxDistance*i), botY-5, gNode.siteList[1].NodeList[i].MyAddress)
				}
				if gNode.numSite >= 3 {
					//left
					painter.Save()
					painter.Translate3(float64(xCoordLeft), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, -(imageHeight + 5), gNode.siteList[2].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
				//right
				if gNode.numSite == 4 {
					painter.Save()
					painter.Translate3(float64(xCoordRight), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, 12, gNode.siteList[3].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
			}
			switch gNode.numSite {
			case 2:
			case 3:
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, rightY), gNode.nodeImageLeft)
				painter.Save()
				painter.Translate3(float64(xCoordLeft), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(-float64(90))
				painter.DrawText3(-(siteNumOffset + 12), 15, "3")
				painter.Restore()
			case 4:
				painter.Save()
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, rightY), gNode.nodeImageLeft)
				painter.Translate3(float64(xCoordLeft), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(-float64(90))
				painter.DrawText3(-(siteNumOffset + 15), 15, "3")
				painter.Restore()
				painter.DrawImage8(core.NewQPoint2(xCoordRight, rightY), gNode.nodeImageRight)
				painter.Save()
				painter.Translate3(float64(xCoordRight), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(float64(90))
				painter.DrawText3(siteNumOffset+15, -17, "4")
				painter.Restore()
			}

			gNode.drawStatus(painter, rightY+5, botY, topY+20, leftY-90, xCoordMid, xCoordLeft, xCoordRight)

		case 5:
			painter.DrawImage8(core.NewQPoint2(xCoordMid, botY), gNode.nodeImageBot) //2 are always drawn
			painter.DrawText3(xCoordMid+siteNumOffset+40, topY+13, "1")
			painter.DrawImage8(core.NewQPoint2(xCoordMid, topY), gNode.nodeImageTop)
			painter.DrawText3(xCoordMid+siteNumOffset+30, botY+26, "2")

			if gNode.drawNodeAddr { //to keep magic numbers easier
				//top
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+(statusBoxDistance*i), topY+imageHeight+10, gNode.siteList[0].NodeList[i].MyAddress)
				}
				//bottom
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+(statusBoxDistance*i), botY-5, gNode.siteList[1].NodeList[i].MyAddress)
				}
				if gNode.numSite >= 3 {
					//left
					painter.Save()
					painter.Translate3(float64(xCoordLeft), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, -(imageHeight + 5), gNode.siteList[2].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
				//right
				if gNode.numSite == 4 {
					painter.Save()
					painter.Translate3(float64(xCoordRight), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, 12, gNode.siteList[3].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
			}
			switch gNode.numSite {
			case 2:
			case 3:
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, rightY), gNode.nodeImageLeft)
				painter.Save()
				painter.Translate3(float64(xCoordLeft), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(-float64(90))
				painter.DrawText3(-(siteNumOffset + 25), 15, "3")
				painter.Restore()
			case 4:
				painter.Save()
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, rightY), gNode.nodeImageLeft)
				painter.Translate3(float64(xCoordLeft), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(-float64(90))
				painter.DrawText3(-(siteNumOffset + 25), 15, "3")
				painter.Restore()
				painter.DrawImage8(core.NewQPoint2(xCoordRight, rightY), gNode.nodeImageRight)
				painter.Save()
				painter.Translate3(float64(xCoordRight), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(float64(90))
				painter.DrawText3(siteNumOffset+43, -17, "4")
				painter.Restore()
			}

			gNode.drawStatus(painter, rightY, botY, topY, leftY-120, xCoordMid, xCoordLeft, xCoordRight)

		case 6:
			painter.DrawImage8(core.NewQPoint2(xCoordMid, botY), gNode.nodeImageBot) //2 are always drawn
			painter.DrawText3(xCoordMid+siteNumOffset*2+10, topY+13, "1")
			painter.DrawImage8(core.NewQPoint2(xCoordMid, topY), gNode.nodeImageTop)
			painter.DrawText3(xCoordMid+siteNumOffset*2+12, botY+26, "2")

			if gNode.drawNodeAddr { //to keep magic numbers easier
				//top
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+(statusBoxDistance*i), topY+imageHeight+10, gNode.siteList[0].NodeList[i].MyAddress)
				}
				//bottom
				for i := 0; i < gNode.numNodePerSite; i++ {
					painter.DrawText3(xCoordMid+(statusBoxDistance*i), botY-5, gNode.siteList[1].NodeList[i].MyAddress)
				}
				if gNode.numSite >= 3 {
					//left
					painter.Save()
					painter.Translate3(float64(xCoordLeft), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, -(imageHeight + 5), gNode.siteList[2].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
				//right
				if gNode.numSite == 4 {
					painter.Save()
					painter.Translate3(float64(xCoordRight), float64(rightY))
					painter.Rotate(float64(90))
					for i := 0; i < gNode.numNodePerSite; i++ {
						painter.DrawText3(statusBoxDistance*i, 12, gNode.siteList[3].NodeList[i].MyAddress)
					}
					painter.Restore()
				}
			}
			switch gNode.numSite {
			case 2:
			case 3:
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, rightY), gNode.nodeImageLeft)
				painter.Save()
				painter.Translate3(float64(xCoordLeft), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(-float64(90))
				painter.DrawText3(-(siteNumOffset + 40), 15, "3")
				painter.Restore()
			case 4:
				painter.Save()
				painter.DrawImage8(core.NewQPoint2(xCoordLeft, rightY), gNode.nodeImageLeft)
				painter.Translate3(float64(xCoordLeft), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(-float64(90))
				painter.DrawText3(-(siteNumOffset + 40), 15, "3")
				painter.Restore()
				painter.DrawImage8(core.NewQPoint2(xCoordRight, rightY), gNode.nodeImageRight)
				painter.Save()
				painter.Translate3(float64(xCoordRight), float64(rightY)) //messing with the coord system is awful
				painter.Rotate(float64(90))
				painter.DrawText3(siteNumOffset+88, -17, "4")
				painter.Restore()
			}
			gNode.drawStatus(painter, rightY, botY, topY, leftY-120, xCoordMid, xCoordLeft, xCoordRight)
		}
		gNode.paintMessages(painter)
	} else if gNode.deadlockDeclared {
		fmt.Println("dedlock paint")
		painter.FillRect3(gNode.frontend.DrawArea.Rect(), gui.NewQBrush3(gui.QColor_FromRgb(16711680), core.Qt__SolidPattern))
		painter.DrawText3(200, 50, "DEADLOCK DETECTED")
	} else {
		painter.FillRect3(event.Rect(), gui.NewQBrush3(gui.QColor_FromRgb(255), core.Qt__SolidPattern))
		painter.Save()
		font := painter.Font()
		font.SetPointSize(font.PointSize() * 2)
		painter.SetFont(font)
		painter.DrawText3(150, 50, "Green square: work ongoing")
		painter.DrawText3(150, 70, "Yellow square: no work, but no issues")
		painter.DrawText3(150, 90, "Black square, initializing or crashed")
		painter.DrawText3(150, 110, "Red square: Deadlock")
		painter.DrawText3(150, 130, "Orange square: sent probe")
		painter.Restore()
	}

	painter.DestroyQPainter()
}
func (gNode *GUINode) drawStatus(painter *gui.QPainter, sideY int, botY int, topY int, leftY int, xCoordMid int, xCoordLeft int, xCoordRight int) {
	//takes all the values used to paint the pictures and uses them to paint the squares
	rectangle := core.NewQRect4(0, 0, 10, 10)
	painter.Save()
	var magicLx = xCoordLeft //all magic numbers, they aren't based off much
	var magicRx = xCoordRight
	var magicLy = leftY
	var magicRy = sideY
	var magicBx = xCoordMid
	var magicTx = xCoordMid
	var magicTy = topY
	var magicBy = botY
	for i, site := range gNode.statusArray {
		for j, node := range site {
			switch node {
			case 0:
				painter.SetBrush(gui.NewQBrush3(gui.NewQColor3(0, 0, 0, 255), core.Qt__SolidPattern))
			case 1:
				painter.SetBrush(gui.NewQBrush3(gui.NewQColor3(255, 255, 0, 255), core.Qt__SolidPattern)) //no work
			case 2:
				painter.SetBrush(gui.NewQBrush3(gui.NewQColor3(0, 255, 0, 255), core.Qt__SolidPattern)) //has work
			case 3:
				painter.SetBrush(gui.NewQBrush3(gui.NewQColor3(255, 165, 0, 255), core.Qt__SolidPattern)) //probe
			case 4:
				painter.SetBrush(gui.NewQBrush3(gui.NewQColor3(255, 0, 0, 255), core.Qt__SolidPattern)) //deadlock
			default:
				painter.SetBrush(gui.NewQBrush3(gui.NewQColor3(0, 0, 125, 255), core.Qt__SolidPattern)) //deadlock
			}
			switch i {
			case 0:
				rectangle = core.NewQRect4(magicTx+50*j, magicTy, 10, 10)
			case 1:
				rectangle = core.NewQRect4(magicBx+50*j, magicBy, 10, 10)
			case 2:
				rectangle = core.NewQRect4(magicLx, magicLy+50*j, 10, 10)
			case 3:
				rectangle = core.NewQRect4(magicRx, magicRy+50*j, 10, 10)
			}
			if gNode.firstRun { //this sets up the positions for painting squares
				gNode.nodePosMap[i][j] = rectangle.Center()
			}
			painter.FillRect3(rectangle, painter.Brush())
		}
	}
	painter.Restore()
	gNode.firstRun = false
}
func (gNode *GUINode) paintMessages(painter *gui.QPainter) { //this doesn't work and I have no idea why
	//its supposed to draw a line between the nodes that are communicating so it looks like a graph
	//the setbrush draw etc methods all work on their own, it entires the loop
	//but it just won't actually display
	painter.Save()
	var tempList []messageRepresentation
	for _, message := range gNode.visibleMessagesList {
		switch message.color {
		case "blue":
			painter.SetBrush(gui.NewQBrush3(gui.NewQColor3(0, 0, 125, 255), core.Qt__SolidPattern)) //deadlock}
			painter.SetPen2(gui.NewQColor6("blue"))
		case "black":
			painter.SetBrush(gui.NewQBrush3(gui.NewQColor3(255, 255, 255, 255), core.Qt__SolidPattern)) //deadlock}
			painter.SetPen2(gui.NewQColor6("black"))
		}
		painter.DrawLine2(core.NewQLine3(message.start.X(), message.start.Y(), message.end.X(), message.end.Y()))
	}
	gNode.visibleMessagesList = tempList
	painter.Restore()
}

func (gNode *GUINode) reset() {
	fmt.Println("reset")

}
