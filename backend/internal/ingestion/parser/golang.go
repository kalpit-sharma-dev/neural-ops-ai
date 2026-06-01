package parser

import (
	"regexp"
	"strings"

	"github.com/neuralops/platform/internal/domain"
)

var (
	goPanicLine  = regexp.MustCompile(`^panic:\s*(.+)$`)
	goFrameLine  = regexp.MustCompile(`^\s*(.+)\((0x[0-9a-f]+)\)\s*$`)
	goFileLine   = regexp.MustCompile(`^\s*(\S+):(\d+)\s*(.*)$`)
)

func parseGolangStackTrace(message string) *domain.StackTrace {
	lines := strings.Split(message, "\n")
	trace := &domain.StackTrace{
		Language: domain.StackTraceLanguageGolang,
		Frames:   []domain.StackFrame{},
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 || strings.HasPrefix(trimmed, "panic:") {
			if match := goPanicLine.FindStringSubmatch(trimmed); match != nil {
				trace.ExceptionType = "panic"
				trace.ExceptionMessage = match[1]
			}
			continue
		}

		if frameMatch := goFrameLine.FindStringSubmatch(trimmed); frameMatch != nil {
			fn := frameMatch[1]
			pkg := fn
			if idx := strings.LastIndex(fn, "."); idx >= 0 {
				pkg = fn[:idx]
				fn = fn[idx+1:]
			}
			trace.Frames = append(trace.Frames, domain.StackFrame{
				Package:      pkg,
				FunctionName: fn,
				FileName:     frameMatch[2],
				IsThirdParty: isThirdParty(pkg),
			})
			continue
		}

		if fileMatch := goFileLine.FindStringSubmatch(trimmed); fileMatch != nil {
			if len(trace.Frames) == 0 {
				continue
			}
			last := len(trace.Frames) - 1
			lineNumber := 0
			if n, err := parseInt(fileMatch[2]); err == nil {
				lineNumber = n
			}
			trace.Frames[last].FileName = fileMatch[1]
			trace.Frames[last].LineNumber = lineNumber
		}
	}

	if trace.ExceptionType == "" && !strings.Contains(message, "goroutine") && len(trace.Frames) == 0 {
		return nil
	}
	if trace.ExceptionType == "" {
		trace.ExceptionType = "runtime error"
		trace.ExceptionMessage = strings.TrimSpace(lines[0])
	}
	return trace
}
