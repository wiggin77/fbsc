package main

import (
	"fbsc/termprinter"

	"github.com/mattermost/logr/v2"
	"github.com/mattermost/logr/v2/formatters"
)

func initLogging(lgr *logr.Logr, printer *termprinter.Printer) error {
	formatter := &formatters.Plain{Delim: "  "}

	filter := &logr.CustomFilter{}
	filter.Add(
		logr.Level{ID: 6, Name: "trace"},
		logr.Level{ID: 5, Name: "debug"},
		logr.Level{ID: 4, Name: "info"},
		logr.Level{ID: 3, Name: "warn"},

		logr.Level{ID: 0, Name: "panic"},
		logr.Level{ID: 1, Name: "fatal"},
		logr.Level{ID: 2, Name: "error", Stacktrace: true},
	)

	t1 := newPrinterTarget(printer)
	return lgr.AddTarget(t1, "stdout", filter, formatter, 1000)
}

// printerTarget is a simple Logr target that writes to the terminal printer.
type printerTarget struct {
	printer *termprinter.Printer
}

func newPrinterTarget(printer *termprinter.Printer) *printerTarget {
	return &printerTarget{
		printer: printer,
	}
}

func (pr *printerTarget) Init() error {
	return nil
}

func (pt *printerTarget) Write(p []byte, rec *logr.LogRec) (int, error) {
	err := pt.printer.Print(string(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (pt *printerTarget) Shutdown() error {
	pt.printer.Erase()
	return nil
}
