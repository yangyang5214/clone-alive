package simple

import (
	"github.com/yangyang5214/clone-alive/pkg/types"
	"os"
	"testing"
)

func TestStatic(t *testing.T) {
	dataStore, _ := os.MkdirTemp("", "clone-alive-*")
	defer func() {
		_ = os.RemoveAll(dataStore)
	}()

	simpleCrawler, err := New(&types.Options{
		TargetDir: dataStore,
	})
	if err != nil {
		panic(err)
	}
	//_, err = simpleCrawler.CrawlAndSave("https://www.baidu.com/", "index.html")
	_, err = simpleCrawler.CrawlAndSave("https://183.134.103.232/resources/jquery/base64.js%3Fjs_ver=2023-02-16", "index.html") //404
	if err != nil {
		t.Log(err)
	}
}
