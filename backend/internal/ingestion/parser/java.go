package parser

import (
	"regexp"
	"strings"

	"github.com/neuralops/platform/internal/domain"
)

var (
	javaExceptionLine = regexp.MustCompile(`^([\w.$]+(?:Exception|Error)):\s*(.*)$`)
	javaFrameLine     = regexp.MustCompile(`^\s*at\s+([\w.$]+)\.([\w$<>]+)\(([\w.$]+):(\d+)\)\s*$`)
)

func parseJavaStackTrace(message string) *domain.StackTrace {
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return nil
	}

	first := strings.TrimSpace(lines[0])
	match := javaExceptionLine.FindStringSubmatch(first)
	if match == nil {
		if !strings.Contains(first, "Exception") && !strings.Contains(first, "Error") {
			return nil
		}
	}

	trace := &domain.StackTrace{
		Language:         domain.StackTraceLanguageJava,
		ExceptionType:    strings.TrimSpace(first),
		ExceptionMessage: "",
		Frames:           []domain.StackFrame{},
	}

	if match != nil {
		trace.ExceptionType = match[1]
		trace.ExceptionMessage = match[2]
	}

	for _, line := range lines[1:] {
		frameMatch := javaFrameLine.FindStringSubmatch(strings.TrimSpace(line))
		if frameMatch == nil {
			continue
		}
		className := frameMatch[1]
		fileName := frameMatch[3]
		lineNumber := 0
		if n, err := parseInt(frameMatch[4]); err == nil {
			lineNumber = n
		}
		trace.Frames = append(trace.Frames, domain.StackFrame{
			ClassName:    className,
			MethodName:   frameMatch[2],
			FileName:     fileName,
			LineNumber:   lineNumber,
			IsThirdParty: isThirdParty(className),
		})
	}

	if len(trace.Frames) == 0 && trace.ExceptionType == "" {
		return nil
	}
	return trace
}
