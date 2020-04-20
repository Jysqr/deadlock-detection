package ui

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

type __deadlockdetectionsimulator struct{}

func (*__deadlockdetectionsimulator) init() {}

type DeadlockDetectionSimulator struct {
	*__deadlockdetectionsimulator
	*widgets.QMainWindow
	ActionQuit       *widgets.QAction
	ActionReset      *widgets.QAction
	ActionInfo       *widgets.QAction
	Centralwidget    *widgets.QWidget
	VerticalLayout_4 *widgets.QVBoxLayout
	Splitter         *widgets.QSplitter
	Numberinfo       *widgets.QWidget
	VerticalLayout   *widgets.QVBoxLayout
	Numlayout        *widgets.QHBoxLayout
	Numpersite       *widgets.QWidget
	VerticalLayout_2 *widgets.QVBoxLayout
	Label            *widgets.QLabel
	SpinBox          *widgets.QSpinBox
	Numsite          *widgets.QWidget
	VerticalLayout_3 *widgets.QVBoxLayout
	Label_2          *widgets.QLabel
	SpinBox_2        *widgets.QSpinBox
	OkLayout         *widgets.QWidget
	VerticalLayout_5 *widgets.QVBoxLayout
	Label_3          *widgets.QLabel
	OkButton         *widgets.QPushButton
	Display          *widgets.QWidget
	Consolelayout    *widgets.QWidget
	HorizontalLayout *widgets.QHBoxLayout
	CheckBox_2       *widgets.QCheckBox
	StartPauseButton *widgets.QPushButton
	HorizontalSpacer *widgets.QSpacerItem
	CheckBox         *widgets.QCheckBox
	Menubar          *widgets.QMenuBar
	MenuFile         *widgets.QMenu
}

func NewDeadlockDetectionSimulator(p widgets.QWidget_ITF) *DeadlockDetectionSimulator {
	var par *widgets.QWidget
	if p != nil {
		par = p.QWidget_PTR()
	}
	w := &DeadlockDetectionSimulator{QMainWindow: widgets.NewQMainWindow(par, 0)}
	w.setupUI()
	w.init()
	return w
}
func (w *DeadlockDetectionSimulator) setupUI() {
	if w.ObjectName() == "" {
		w.SetObjectName("DeadlockDetectionSimulator")
	}
	w.Resize2(854, 665)
	w.ActionQuit = widgets.NewQAction(w)
	w.ActionQuit.SetObjectName("actionQuit")
	w.ActionReset = widgets.NewQAction(w)
	w.ActionReset.SetObjectName("actionReset")
	w.ActionInfo = widgets.NewQAction(w)
	w.ActionInfo.SetObjectName("actionInfo")
	w.Centralwidget = widgets.NewQWidget(w, 0)
	w.Centralwidget.SetObjectName("centralwidget")
	w.VerticalLayout_4 = widgets.NewQVBoxLayout2(w.Centralwidget)
	w.VerticalLayout_4.SetObjectName("verticalLayout_4")
	w.Splitter = widgets.NewQSplitter(w.Centralwidget)
	w.Splitter.SetObjectName("splitter")
	w.Splitter.SetOrientation(core.Qt__Vertical)
	w.Numberinfo = widgets.NewQWidget(w.Splitter, 0)
	w.Numberinfo.SetObjectName("numberinfo")
	w.Numberinfo.SetEnabled(true)
	w.Numberinfo.SetMinimumSize(core.NewQSize2(0, 0))
	w.VerticalLayout = widgets.NewQVBoxLayout2(w.Numberinfo)
	w.VerticalLayout.SetSpacing(0)
	w.VerticalLayout.SetObjectName("verticalLayout")
	w.VerticalLayout.SetContentsMargins(0, 0, 0, 0)
	w.Numlayout = widgets.NewQHBoxLayout()
	w.Numlayout.SetSpacing(3)
	w.Numlayout.SetObjectName("numlayout")
	w.Numpersite = widgets.NewQWidget(w.Numberinfo, 0)
	w.Numpersite.SetObjectName("numpersite")
	w.VerticalLayout_2 = widgets.NewQVBoxLayout2(w.Numpersite)
	w.VerticalLayout_2.SetSpacing(0)
	w.VerticalLayout_2.SetObjectName("verticalLayout_2")
	w.VerticalLayout_2.SetContentsMargins(0, 0, 0, 0)
	w.Label = widgets.NewQLabel(w.Numpersite, 0)
	w.Label.SetObjectName("label")
	w.VerticalLayout_2.QLayout.AddWidget(w.Label)
	w.SpinBox = widgets.NewQSpinBox(w.Numpersite)
	w.SpinBox.SetObjectName("spinBox")
	w.SpinBox.SetMinimum(2)
	w.SpinBox.SetMaximum(4)
	w.VerticalLayout_2.QLayout.AddWidget(w.SpinBox)
	w.Numlayout.QLayout.AddWidget(w.Numpersite)
	w.Numsite = widgets.NewQWidget(w.Numberinfo, 0)
	w.Numsite.SetObjectName("numsite")
	w.VerticalLayout_3 = widgets.NewQVBoxLayout2(w.Numsite)
	w.VerticalLayout_3.SetSpacing(0)
	w.VerticalLayout_3.SetObjectName("verticalLayout_3")
	w.VerticalLayout_3.SetContentsMargins(0, 0, 0, 0)
	w.Label_2 = widgets.NewQLabel(w.Numsite, 0)
	w.Label_2.SetObjectName("label_2")
	w.VerticalLayout_3.QLayout.AddWidget(w.Label_2)
	w.SpinBox_2 = widgets.NewQSpinBox(w.Numsite)
	w.SpinBox_2.SetObjectName("spinBox_2")
	w.SpinBox_2.SetMinimum(2)
	w.SpinBox_2.SetMaximum(6)
	w.VerticalLayout_3.QLayout.AddWidget(w.SpinBox_2)
	w.Numlayout.QLayout.AddWidget(w.Numsite)
	w.OkLayout = widgets.NewQWidget(w.Numberinfo, 0)
	w.OkLayout.SetObjectName("okLayout")
	w.VerticalLayout_5 = widgets.NewQVBoxLayout2(w.OkLayout)
	w.VerticalLayout_5.SetSpacing(0)
	w.VerticalLayout_5.SetObjectName("verticalLayout_5")
	w.VerticalLayout_5.SetContentsMargins(0, 0, 0, 0)
	w.Label_3 = widgets.NewQLabel(w.OkLayout, 0)
	w.Label_3.SetObjectName("label_3")
	w.VerticalLayout_5.QLayout.AddWidget(w.Label_3)
	w.OkButton = widgets.NewQPushButton(w.OkLayout)
	w.OkButton.SetObjectName("okButton")
	sizePolicy := widgets.NewQSizePolicy2(widgets.QSizePolicy__Fixed, widgets.QSizePolicy__Fixed, 0)
	sizePolicy.SetHorizontalStretch(0)
	sizePolicy.SetVerticalStretch(0)
	sizePolicy.SetHeightForWidth(w.OkButton.SizePolicy().HasHeightForWidth())
	w.OkButton.SetSizePolicy(sizePolicy)
	w.OkButton.SetMaximumSize(core.NewQSize2(16777210, 16777215))
	w.VerticalLayout_5.QLayout.AddWidget(w.OkButton)
	w.Numlayout.QLayout.AddWidget(w.OkLayout)
	w.VerticalLayout.AddLayout(w.Numlayout, 0)
	w.Splitter.AddWidget(w.Numberinfo)
	w.Display = widgets.NewQWidget(w.Splitter, 0)
	w.Display.SetObjectName("display")
	sizePolicy1 := widgets.NewQSizePolicy2(widgets.QSizePolicy__Minimum, widgets.QSizePolicy__Minimum, 0)
	sizePolicy1.SetHorizontalStretch(0)
	sizePolicy1.SetVerticalStretch(0)
	sizePolicy1.SetHeightForWidth(w.Display.SizePolicy().HasHeightForWidth())
	w.Display.SetSizePolicy(sizePolicy1)
	w.Display.SetMinimumSize(core.NewQSize2(500, 500))
	w.Display.SetMouseTracking(true)
	w.Display.SetAutoFillBackground(false)
	w.Splitter.AddWidget(w.Display)
	w.Consolelayout = widgets.NewQWidget(w.Splitter, 0)
	w.Consolelayout.SetObjectName("consolelayout")
	w.HorizontalLayout = widgets.NewQHBoxLayout2(w.Consolelayout)
	w.HorizontalLayout.SetSpacing(0)
	w.HorizontalLayout.SetObjectName("horizontalLayout")
	w.HorizontalLayout.SetContentsMargins(0, 0, 0, 0)
	w.CheckBox_2 = widgets.NewQCheckBox(w.Consolelayout)
	w.CheckBox_2.SetObjectName("checkBox_2")
	w.HorizontalLayout.QLayout.AddWidget(w.CheckBox_2)
	w.StartPauseButton = widgets.NewQPushButton(w.Consolelayout)
	w.StartPauseButton.SetObjectName("startPauseButton")
	w.HorizontalLayout.QLayout.AddWidget(w.StartPauseButton)
	w.HorizontalSpacer = widgets.NewQSpacerItem(603, 20, widgets.QSizePolicy__Expanding, widgets.QSizePolicy__Minimum)
	w.HorizontalLayout.AddItem(w.HorizontalSpacer)
	w.CheckBox = widgets.NewQCheckBox(w.Consolelayout)
	w.CheckBox.SetObjectName("checkBox")
	w.HorizontalLayout.QLayout.AddWidget(w.CheckBox)
	w.Splitter.AddWidget(w.Consolelayout)
	w.VerticalLayout_4.QLayout.AddWidget(w.Splitter)
	w.SetCentralWidget(w.Centralwidget)
	w.Menubar = widgets.NewQMenuBar(w)
	w.Menubar.SetObjectName("menubar")
	w.Menubar.SetGeometry(core.NewQRect4(0, 0, 854, 28))
	w.MenuFile = widgets.NewQMenu(w.Menubar)
	w.MenuFile.SetObjectName("menuFile")
	w.SetMenuBar(w.Menubar)
	if true {
		w.Label.SetBuddy(w.SpinBox)
		w.Label_2.SetBuddy(w.SpinBox_2)
		w.Label_3.SetBuddy(w.OkButton)
	}
	widgets.QWidget_SetTabOrder(w.SpinBox, w.SpinBox_2)
	widgets.QWidget_SetTabOrder(w.SpinBox_2, w.CheckBox_2)
	widgets.QWidget_SetTabOrder(w.CheckBox_2, w.CheckBox)
	widgets.QWidget_SetTabOrder(w.CheckBox, w.OkButton)
	widgets.QWidget_SetTabOrder(w.OkButton, w.StartPauseButton)
	w.Menubar.AddActions([]*widgets.QAction{w.MenuFile.MenuAction()})
	w.MenuFile.AddActions([]*widgets.QAction{w.ActionInfo})
	w.MenuFile.AddActions([]*widgets.QAction{w.ActionReset})
	w.MenuFile.AddActions([]*widgets.QAction{w.ActionQuit})
	w.retranslateUi()
	w.StartPauseButton.SetDefault(false)
	core.QMetaObject_ConnectSlotsByName(w)

}
func (w *DeadlockDetectionSimulator) retranslateUi() {
	w.SetWindowTitle(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "Deadlock Detection Simulator", "", 0))
	w.ActionQuit.SetText(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "Quit", "", 0))
	w.ActionReset.SetText(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "Reset", "", 0))
	w.ActionInfo.SetText(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "Info", "", 0))
	w.Label.SetText(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "Number of Sites", "", 0))
	w.Label_2.SetText(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "Number of Nodes per Site", "", 0))
	w.Label_3.SetText("")
	w.OkButton.SetText(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "OK", "", 0))
	w.CheckBox_2.SetText(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "Step", "", 0))
	w.StartPauseButton.SetText(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "Start", "", 0))
	w.CheckBox.SetText(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "Show Console", "", 0))
	w.MenuFile.SetTitle(core.QCoreApplication_Translate("DeadlockDetectionSimulator", "File", "", 0))

}
