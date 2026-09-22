package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

type level int

const (
	levelDebug level = iota
	levelInfo
	levelWarn
	levelError
)

var currentLevel = levelInfo

// Init sets the minimum log level. Accepted values: "debug", "info", "warn", "error".
// Unrecognised values fall back to "info".
func Init(lvl string) {
	switch strings.ToLower(lvl) {
	case "debug":
		currentLevel = levelDebug
	case "warn":
		currentLevel = levelWarn
	case "error":
		currentLevel = levelError
	default:
		currentLevel = levelInfo
	}
}

// Info logs an informational message (startup progress, normal operations).
func Info(format string, v ...any) {
	if currentLevel <= levelInfo {
		writeLog("[INFO]", format, v...)
	}
}

// Debug logs a verbose diagnostic message (request tracing, path details).
func Debug(format string, v ...any) {
	if currentLevel <= levelDebug {
		writeLog("[DEBUG]", format, v...)
	}
}

// Warn logs a warning that the application can recover from.
func Warn(format string, v ...any) {
	if currentLevel <= levelWarn {
		writeLog("[WARN]", format, v...)
	}
}

// Error logs an error condition.
func Error(format string, v ...any) {
	if currentLevel <= levelError {
		writeLog("[ERROR]", format, v...)
	}
}

// Fatal logs an error then exits with status 1.
func Fatal(format string, v ...any) {
	writeLog("[ERROR]", format, v...)
	os.Exit(1)
}

func writeLog(severity, format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	log.Printf("%s %s", severity, neutralizeLogText(message))
}

func neutralizeLogText(message string) string {
	var safe strings.Builder
	for _, character := range message {
		switch character {
		case '\r':
			safe.WriteString(`\r`)
		case '\n':
			safe.WriteString(`\n`)
		case '\u2028':
			safe.WriteString(`\u2028`)
		case '\u2029':
			safe.WriteString(`\u2029`)
		default:
			if unicode.IsControl(character) {
				if character <= 0xff {
					fmt.Fprintf(&safe, `\x%02x`, character)
				} else {
					fmt.Fprintf(&safe, `\u%04x`, character)
				}
				continue
			}
			safe.WriteRune(character)
		}
	}
	return safe.String()
}
