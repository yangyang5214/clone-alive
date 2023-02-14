package utils

import (
	"fmt"
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
	return ExecCommand(cmdStr)
}

func ExecCommand(cmd string) string {
	output, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(output), "\n")
}
