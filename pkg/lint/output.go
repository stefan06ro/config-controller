package lint

import (
	"fmt"
	"sort"
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
