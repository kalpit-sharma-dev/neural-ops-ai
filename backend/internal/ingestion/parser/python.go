package parser

import (
	"regexp"
	"strings"

	"github.com/neuralops/platform/internal/domain"
)

var (
	pythonTracebackLine = regexp.MustCompile(`^Traceback \(most recent call last\):`)
	pythonFrameLine     = regexp.MustCompile(`^\s*File "(.+)", line (\d+), in (.+)$`)
	pythonErrorLine     = regexp.MustCompile(`^([\w]+(?:Error|Exception)):\s*(.*)$`)
)

func parsePythonStackTrace(message string) *domain.StackTrace {
	if !pythonTracebackLine.MatchString(message) && !strings.Contains(message, "Traceback") {
		return nil
	}

	lines := strings.Split(message, "\n")
	trace := &domain.StackTrace{
		Language: domain.StackTraceLanguagePython,
		Frames:   []domain.StackFrame{},
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if match := pythonFrameLine.FindStringSubmatch(trimmed); match != nil {
			lineNumber := 0
			if n, err := parseInt(match[2]); err == nil {
				lineNumber = n
			}
			trace.Frames = append(trace.Frames, domain.StackFrame{
				FileName:     match[1],
				FunctionName: match[3],
				LineNumber:   lineNumber,
				IsThirdParty: isThirdParty(match[1]),
			})
			continue
		}
		if match := pythonErrorLine.FindStringSubmatch(trimmed); match != nil {
			trace.ExceptionType = match[1]
			trace.ExceptionMessage = match[2]
		}
	}

	if len(trace.Frames) == 0 && trace.ExceptionType == "" {
		return nil
	}
	return trace
}
