package output

import (
	"bufio"
	"os"
)

//https://github.com/projectdiscovery/katana/blob/main/pkg/output/file_writer.go

// FileWriter is a concurrent file based output writer.
type FileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

// NewFileOutputWriter creates a new buffered writer for a file
func newFileOutputWriter(file string) (*FileWriter, error) {
	output, err := os.Create(file)
	if err != nil {
		return nil, err
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
	_ = w.writer.Flush()
	//nolint:errcheck // we don't care whether sync failed or succeeded.
	_ = w.file.Sync()
	return w.file.Close()
}
