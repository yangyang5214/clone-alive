package chrome

import (
	"github.com/yangyang5214/clone-alive/pkg/types"
	"os"
	"testing"
)

// TestKillAllChrome is kill all running chromium
func TestKillAllChrome(t *testing.T) {
	chrome, err := New(&types.Options{
		Url: "https://www.baidu.com/",
	})
	defer os.RemoveAll(chrome.targetDir)
	if err != nil {
		panic(err)
	}
	chrome.previousPIDs = nil
	_ = chrome.killChromeProcesses()
}
