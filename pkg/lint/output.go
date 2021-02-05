package lint

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
)

var (
	red  = color.New(color.FgRed).SprintFunc()
	blue = color.New(color.FgBlue).SprintFunc()
)

type LinterMessage interface {
	fmt.Stringer
}

type LinterMessages []LinterMessage

type linterMessage struct {
	sourceFile  string
	path        string
	message     string
	description string // optional
}

type Error struct {
	linterMessage
}

type Suggestion struct {
	linterMessage
}

func (e Error) IsError() bool {
	return true
}

func (e Error) String() string {
	newlineDescription := ""
	if e.description != "" {
		newlineDescription = "\n" + e.description
	}
	return fmt.Sprintf(
		"%s: .%s %s %s",
		red(e.sourceFile),
		red(e.path),
		e.message,
		newlineDescription,
	)
}

func (s Suggestion) String() string {
	newlineDescription := ""
	if s.description != "" {
		newlineDescription = "\n" + s.description
	}
	return fmt.Sprintf(
		"%s: .%s %s %s",
		blue(s.sourceFile),
		blue(s.path),
		s.message,
		newlineDescription,
	)
}

func (s Suggestion) IsError() bool {
	return false
}

// LinterMessages implements sort.Interface
func (m LinterMessages) Len() int {
	return len(m)
}

func (m LinterMessages) Less(i, j int) bool {
	switch {
	case m[i].IsError() == m[j].IsError():
		switch {
		case m[i].sourceFile == m[j].sourceFile:
			return m[i].path <= m[j].path
		default:
			return m[i].sourceFile <= m[j].sourceFile
		}
	case m[i].IsError() && !m[j].IsError():
		return true
	case !m[i].IsError() && m[j].IsError():
		return false
	}
}

func (m LinterMessages) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
