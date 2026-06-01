package parser

import (
	"regexp"
	"strings"

	"github.com/neuralops/platform/internal/domain"
)

var (
	nodeErrorLine = regexp.MustCompile(`^(\w+Error):\s*(.*)$`)
	nodeFrameLine = regexp.MustCompile(`^\s*at\s+(.+?) \((.+):(\d+):(\d+)\)\s*$`)
)

func parseNodeStackTrace(message string) *domain.StackTrace {
	if !strings.Contains(message, "Error") && !strings.Contains(message, "at ") {
		return nil
	}

	lines := strings.Split(message, "\n")
	trace := &domain.StackTrace{
		Language: domain.StackTraceLanguageNodeJS,
		Frames:   []domain.StackFrame{},
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 {
			if match := nodeErrorLine.FindStringSubmatch(trimmed); match != nil {
				trace.ExceptionType = match[1]
				trace.ExceptionMessage = match[2]
			} else if strings.Contains(trimmed, "Error") {
				trace.ExceptionType = trimmed
			}
			continue
		}

		if match := nodeFrameLine.FindStringSubmatch(trimmed); match != nil {
			lineNumber := 0
			if n, err := parseInt(match[3]); err == nil {
				lineNumber = n
			}
			trace.Frames = append(trace.Frames, domain.StackFrame{
				MethodName:   match[1],
				FileName:     match[2],
				LineNumber:   lineNumber,
				IsThirdParty: isThirdParty(match[2]),
			})
		}
	}

	if len(trace.Frames) == 0 && trace.ExceptionType == "" {
		return nil
	}
	return trace
}
