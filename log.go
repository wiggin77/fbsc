package main

import (
	"fmt"
	"os"

	"github.com/mattermost/logr"
	"github.com/mattermost/logr/format"
	"github.com/mattermost/logr/target"
)

func initLogging(lgr *logr.Logr, logConfig string) error {
	if logConfig != "" {
		return initLoggingFromFile(lgr, logConfig)
	}
	return initLoggingDefaults(lgr)
}

func initLoggingDefaults(lgr *logr.Logr) error {
	formatter := &format.Plain{Delim: "  "}

	filter := &logr.CustomFilter{}
	filter.Add(
		logr.Level{ID: 6, Name: "trace"},
		logr.Level{ID: 5, Name: "debug"},
		logr.Level{ID: 4, Name: "info"},
		logr.Level{ID: 3, Name: "warn"},
	)

	t1 := target.NewWriterTarget(filter, formatter, os.Stdout, 1000)

	filter = &logr.CustomFilter{}
	filter.Add(
		logr.Level{ID: 0, Name: "panic"},
		logr.Level{ID: 1, Name: "fatal"},
		logr.Level{ID: 2, Name: "error", Stacktrace: true},
	)

	t2 := target.NewWriterTarget(filter, formatter, os.Stderr, 1000)

	return lgr.AddTarget(t1, t2)
}

func initLoggingFromFile(logr *logr.Logr, logConfig string) error {
	return fmt.Errorf("not implemented")
}
