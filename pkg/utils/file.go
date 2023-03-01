package utils

import (
	"fmt"
	"github.com/projectdiscovery/gologger"
	"os"
	"os/exec"
	"strings"
)

// CurrentDirectory get the current working directory
func CurrentDirectory() string {
	path, _ := os.Getwd()
	return path
}

func ReadFile(filePath string) ([]byte, error) {
	f, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func IsFileExist(p string) bool {
	f, err := os.Stat(p)
	if err != nil {
		return false
	}
	if f.IsDir() {
		return false
	}
	return true
}

func FindFileByName(dir string, name string) string {
	cmdStr := fmt.Sprintf("find %s -name %s", dir, name)
	gologger.Info().Msgf("find cmd <%s>", cmdStr)
	output, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		gologger.Error().Msgf("find error %s", err)
		return ""
	}
	result := strings.Split(string(output), "\n")
	return result[0]
}
