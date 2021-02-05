package lint

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	red  = color.New(color.FgRed).SprintFunc()
	blue = color.New(color.FgBlue).SprintFunc()
)

type LinterMessages []LinterMessage

type LinterMessage struct {
	isError     bool
	sourceFile  string
	path        string
	message     string
	description string // optional
}

func (lm LinterMessage) IsError() bool {
	return lm.isError
}

func (lm LinterMessage) String() string {
	newlineDescription := ""
	if lm.description != "" {
		newlineDescription = "\n" + lm.description
	}

	outputColor := blue
	if lm.isError {
		outputColor = red
	}

	return fmt.Sprintf(
		"%s: .%s %s %s",
		outputColor(lm.sourceFile),
		outputColor(lm.path),
		lm.message,
		newlineDescription,
	)
}

// LinterMessages implements sort.Interface
func (m LinterMessages) Len() int {
	return len(m)
}

func (m LinterMessages) Less(i, j int) bool {
	switch {
	case m[i].isError == m[j].isError:
		switch {
		case m[i].sourceFile == m[j].sourceFile:
			return m[i].path <= m[j].path
		default:
			return m[i].sourceFile <= m[j].sourceFile
		}
	case m[i].isError && !m[j].isError:
		return true
	case !m[i].isError && m[j].isError:
		return false
	default: // this should never happen
		return false
	}
}

func (m LinterMessages) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
