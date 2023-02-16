package output

import (
	"bufio"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"os"
)

//https://github.com/projectdiscovery/katana/blob/main/pkg/output/file_writer.go

// FileWriter is a concurrent file based output writer.
type FileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

// NewFileOutputWriter creates a new buffered writer for a file
func newFileOutputWriter(file string) (fw *FileWriter, err error) {
	var output *os.File
	if utils.IsFileExist(file) {
		//output, err = os.Open(file)
		//https://stackoverflow.com/a/33852107/15188179
		output, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	} else {
		output, err = os.Create(file)
	}
	if err != nil {
		return fw, err
	}
	return &FileWriter{file: output, writer: bufio.NewWriter(output)}, nil
}

// WriteString writes an output to the underlying file
func (w *FileWriter) Write(data []byte) error {
	_, err := w.writer.Write(data)
	if err != nil {
		return err
	}
	_, err = w.writer.WriteRune('\n')
	return err
}

// Close closes the underlying writer flushing everything to disk
func (w *FileWriter) Close() error {
	err := w.writer.Flush()
	if err != nil {
		gologger.Error().Msgf("writer flush error: %s", err.Error())
	}
	return w.file.Close()
}
