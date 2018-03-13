package main

import (
	"os"

	"github.com/elastic/beats/libbeat/cmd"

	"github.com/axw/pprofbeat/beater"
)

var (
	rootCmd = cmd.GenRootCmd("pprofbeat", "", beater.New)
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
