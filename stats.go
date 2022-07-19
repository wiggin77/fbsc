package main

import "fmt"

type stats struct {
	UserCount    int
	ChannelCount int
	BoardCount   int
	ViewCount    int
	CardCount    int
	TextCount    int
}

func (s *stats) Add(s2 stats) {
	s.UserCount += s2.UserCount
	s.ChannelCount += s2.ChannelCount
	s.BoardCount += s2.BoardCount
	s.ViewCount += s2.ViewCount
	s.CardCount += s2.CardCount
	s.TextCount += s2.TextCount
}

// Print prints the stats to std.out and returns the number of lines output.
func (s *stats) PrintLines() []string {
	lines := make([]string, 0, 6)

	var line string
	line = fmt.Sprintf("   Users: %d", s.UserCount)
	lines = append(lines, line)
	line = fmt.Sprintf("Channels: %d", s.ChannelCount)
	lines = append(lines, line)
	line = fmt.Sprintf("  Boards: %d", s.BoardCount)
	lines = append(lines, line)
	line = fmt.Sprintf("   Views: %d", s.ViewCount)
	lines = append(lines, line)
	line = fmt.Sprintf("   Cards: %d", s.CardCount)
	lines = append(lines, line)
	line = fmt.Sprintf("    Text: %d", s.TextCount)
	lines = append(lines, line)

	return lines
}
