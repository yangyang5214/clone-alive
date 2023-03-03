package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func FindFileByName(dir string, name string) string {
	cmdStr := fmt.Sprintf("find %s -name %s", dir, name)
	output, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		return ""
	}
	result := strings.Split(string(output), "\n")
	return result[0]
}
