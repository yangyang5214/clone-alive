package chrome

import (
	"os"
	"testing"

	"github.com/yangyang5214/clone-alive/pkg/types"
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
