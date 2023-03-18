package output

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"path/filepath"
)

var RouterFile = "runtime.log"

// Writer is an interface which writes output to somewhere for katana events.
type Writer interface {
	Close() error
	Write(types.ResponseResult) error
}

type StandardWriter struct {
	outputFile *FileWriter
}

// New creates a new StandardWriter obj
func New(targetDir string) (Writer, error) {
	writer := &StandardWriter{}
	output, err := newFileOutputWriter(filepath.Join(targetDir, RouterFile))
	if err != nil {
		return nil, errors.Wrap(err, "could not create output file")
	}
	writer.outputFile = output
	return writer, nil
}

func (s *StandardWriter) Write(respResult types.ResponseResult) error {
	respResult.BodyLen = len(respResult.Body)
	respResult.Body = ""
	data, err := json.Marshal(&respResult)
	if err != nil {
		return errors.Wrap(err, "json Marshal error")
	}
	gologger.Info().Msg(string(data))
	err = s.outputFile.Write(data)
	if err != nil {
		gologger.Error().Msgf("write file error, %s", err.Error())
		return err
	}
	return nil
}

func (s *StandardWriter) Close() error {
	err := s.outputFile.Close()
	if err != nil {
		return err
	}
	return nil
}
