package ui

import (
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
The code in uic_gui is generated from the gui_frontend.ui file found within this directory.
This file will add all the actual functions and fun stuff to the generated code as the generated code is just the pretty parts.
*/
type DeadlockSimulatorGUI struct {
	frontend *DeadlockDetectionSimulator
	app      *widgets.QApplication
}

func NewDeadlockSimulatorGUI() *DeadlockSimulatorGUI {
	tempApp := widgets.NewQApplication(len(os.Args), os.Args)
	ui := NewDeadlockDetectionSimulator(nil)

	dls := &DeadlockSimulatorGUI{
		frontend: ui,
		app:      tempApp,
	}
	return dls
}

func (gui *DeadlockSimulatorGUI) ok() {

}
