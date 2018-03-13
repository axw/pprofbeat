package fetch

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/axw/pprofbeat/internal/pprof"
	"github.com/google/pprof/driver"
	"github.com/google/pprof/profile"
)

const (
	outputName = "output"
)

type Options struct {
	Duration time.Duration
	Timeout  time.Duration
	URL      string
}

func (o Options) args() []string {
	if o.URL == "" {
		panic("URL not specified")
	}
	args := []string{
		"-proto", // output compressed protobuf
		"-output", outputName,
		"-symbolize=remote",
	}
	if o.Timeout > 0 {
		args = append(args, "-timeout", fmt.Sprint(o.Timeout.Seconds()))
	}
	if o.Duration > 0 {
		args = append(args, "-seconds", fmt.Sprint(o.Duration.Seconds()))
	}
	args = append(args, o.URL)
	return args
}

// TODO(axw) comment
func Fetch(opts Options) (*profile.Profile, error) {
	// Create a temporary directory in which the profile
	// will be stored. Otherwise we'll leave junk behind.
	tempdir, err := ioutil.TempDir("", "pprofbeat")
	if err != nil {
		return nil, err
	}
	os.Setenv("PPROF_TMPDIR", tempdir)
	defer os.Unsetenv("PPROF_TMPDIR")
	defer os.RemoveAll(tempdir)

	flagSet := &pprofFlagSet{
		FlagSet: flag.NewFlagSet("pprof", flag.PanicOnError),
		args:    opts.args(),
	}
	var w inmemWriter
	if err := driver.PProf(&driver.Options{
		Fetch:   &pprof.Fetcher{},
		Flagset: flagSet,
		Writer:  &w,
		UI:      nonInteractive{},
	}); err != nil {
		return nil, err
	}
	return profile.Parse(&w)
}

type inmemWriter struct {
	bytes.Buffer
}

func (w *inmemWriter) Open(name string) (io.WriteCloser, error) {
	if name != outputName {
		panic("unexpected file name: " + name)
	}
	return w, nil
}

func (w *inmemWriter) Close() error {
	return nil
}

// TODO(axw) supply a logger for logging Print/PrintErr
type nonInteractive struct{}

func (nonInteractive) ReadLine(prompt string) (string, error) {
	return "", errors.New("readline not possible")
}
func (nonInteractive) Print(...interface{})                {}
func (nonInteractive) PrintErr(...interface{})             {}
func (nonInteractive) IsTerminal() bool                    { return false }
func (nonInteractive) SetAutoComplete(func(string) string) {}
