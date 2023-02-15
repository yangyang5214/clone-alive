package simple

import (
	"crypto/tls"
	"github.com/projectdiscovery/gologger"
	"io"
	"net/http"
	"os"
	"time"
)

type Crawler struct {
	httpClient *http.Client
}

// New is created new Crawler
func New() (*Crawler, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Crawler{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   60 * time.Second},
	}, nil
}

// Crawl crawls a URL with the specified options
func (c *Crawler) Crawl(url string, targetPath string) error {
	gologger.Info().Msgf("For url <%s> => <%s>", url, targetPath)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return nil //这里根据需要忽略
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = os.WriteFile(targetPath, bytes, 0644)
	if err == nil {
		return err
	}
	return nil
}
