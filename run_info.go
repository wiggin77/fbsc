package main

import (
	"fbsc/termprinter"
	"sync"
	"sync/atomic"

	"github.com/mattermost/logr/v2"
)

type runInfo struct {
	cfg        *Config
	printer    *termprinter.Printer
	logger     logr.Logger
	abort      chan struct{}
	admin      *AdminClient
	quiet      bool
	blockCount int64

	statsMux sync.Mutex
	stats    *stats
}

func (ri *runInfo) IncBlockCount(add int) {
	_ = atomic.AddInt64(&ri.blockCount, int64(add))
}

func (ri *runInfo) AddStats(stats stats) {
	ri.statsMux.Lock()
	defer ri.statsMux.Unlock()

	ri.stats.Add(stats)
	ri.printer.UpdateLines(ri.stats.PrintLines())
}
