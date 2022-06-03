package main

import (
	"fmt"
	"os"

	"github.com/mattermost/logr/v2"
	"github.com/mattermost/logr/v2/formatters"
	"github.com/mattermost/logr/v2/targets"
)

func initLogging(lgr *logr.Logr, logConfig string) error {
	if logConfig != "" {
		return initLoggingFromFile(lgr, logConfig)
	}
	return initLoggingDefaults(lgr)
}

func initLoggingDefaults(lgr *logr.Logr) error {
	formatter := &formatters.Plain{Delim: "  "}

	filter := &logr.CustomFilter{}
	filter.Add(
		logr.Level{ID: 6, Name: "trace"},
		logr.Level{ID: 5, Name: "debug"},
		logr.Level{ID: 4, Name: "info"},
		logr.Level{ID: 3, Name: "warn"},
	)

	t1 := targets.NewWriterTarget(os.Stdout)
	if err := lgr.AddTarget(t1, "stdout", filter, formatter, 1000); err != nil {
		return err
	}

	filter = &logr.CustomFilter{}
	filter.Add(
		logr.Level{ID: 0, Name: "panic"},
		logr.Level{ID: 1, Name: "fatal"},
		logr.Level{ID: 2, Name: "error", Stacktrace: true},
	)

	t2 := targets.NewWriterTarget(os.Stderr)
	return lgr.AddTarget(t2, "stderr", filter, formatter, 1000)
}

func initLoggingFromFile(logr *logr.Logr, logConfig string) error {
	return fmt.Errorf("not implemented")
}
