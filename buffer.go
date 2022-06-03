package main

import (
	"strings"
	"sync"
)

type buffer struct {
	buf strings.Builder
	mux sync.Mutex
}

func (b *buffer) Write(s string) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.buf.WriteString(s)
}

func (b *buffer) String() string {
	b.mux.Lock()
	defer b.mux.Unlock()

	return b.buf.String()
}
