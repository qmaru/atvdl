package utils

import (
	"github.com/qmaru/minitools"
)

// AESSuite 初始化
var AESSuite *minitools.AESSuiteBasic

// FileSuite 初始化
var FileSuite *minitools.FileSuiteBasic

func init() {
	AESSuite = minitools.AESSuite()
	FileSuite = minitools.FileSuite()
}
