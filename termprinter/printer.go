package termprinter

import (
	"sync"

	ansi "github.com/k0kubun/go-ansi"
)

type Printer struct {
	mux   sync.Mutex
	lines []string
}

// NewPrinter creates a new terminal printer.
func NewPrinter() *Printer {
	return &Printer{}
}

func (p *Printer) erase() {
	if len(p.lines) == 0 {
		return
	}

	ansi.CursorPreviousLine(len(p.lines))

	for i := 0; i < len(p.lines); i++ {
		ansi.EraseInLine(2)
		ansi.CursorNextLine(1)
	}
	p.lines = nil
}

// UpdateLines prints the specified lines to the terminal. Subsequent calls to UpdatesLines causes
// the lines to be overwritten in-place.
func (p *Printer) UpdateLines(lines []string) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if len(lines) != len(p.lines) {
		p.erase()
	} else {
		ansi.CursorPreviousLine(len(p.lines))
	}

	for _, line := range lines {
		ansi.EraseInLine(2)
		if _, err := ansi.Println(line); err != nil {
			panic(err)
		}
	}

	p.lines = lines
}

// Erase erases any lines added via UpdateLines and leaves the cursor at the beginning of
// the erased block.
func (p *Printer) Erase() {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.erase()
}

func (p *Printer) Print(a ...any) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.erase()

	_, err := ansi.Print(a...)
	return err
}

func (p *Printer) Println(a ...any) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.erase()

	_, err := ansi.Println(a...)
	return err
}

func (p *Printer) Printf(format string, a ...any) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.erase()

	_, err := ansi.Printf(format, a...)
	return err
}
