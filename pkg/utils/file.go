package utils

import "os"

// CurrentDirectory get the current working directory
func CurrentDirectory() string {
	path, _ := os.Getwd()
	return path
}
