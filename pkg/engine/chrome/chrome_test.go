package chrome

import (
	"clone-alive/pkg/types"
	"os"
	"testing"
)

//TestKillAllChrome is kill all running chromium
func TestKillAllChrome(t *testing.T) {
	chrome, _ := New(&types.Options{})
	chrome.previousPIDs = nil
	_ = chrome.killChromeProcesses()
	_ = os.RemoveAll(chrome.targetDir)
}
