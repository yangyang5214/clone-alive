package utils

import "os"

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
