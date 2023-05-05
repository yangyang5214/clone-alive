package utils

import (
	"fmt"
	"github.com/projectdiscovery/gologger"
	"os/exec"
	"strings"
)

func FindFileByName(dir string, name string) string {
	//find bin/mail.chinacdc.com/coremail/bundle -name '$login.e8c17.js' => good luck
	//find bin/mail.chinacdc.com/coremail/bundle -name $login.e8c17.js
	cmdStr := fmt.Sprintf("find %s -name '%s'", dir, name)
	gologger.Debug().Msgf("find cmd: %s", cmdStr)
	output, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		return ""
	}
	result := strings.Split(string(output), "\n")
	return result[0]
}
