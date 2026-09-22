package logger

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestLogLevelsNeutralizeControlCharacters(t *testing.T) {
	var output bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	originalLevel := currentLevel
	log.SetOutput(&output)
	log.SetFlags(0)
	log.SetPrefix("")
	Init("debug")
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
		currentLevel = originalLevel
	})

	levels := []struct {
		name     string
		severity string
		write    func(string, ...any)
	}{
		{name: "debug", severity: "[DEBUG]", write: Debug},
		{name: "info", severity: "[INFO]", write: Info},
		{name: "warn", severity: "[WARN]", write: Warn},
		{name: "error", severity: "[ERROR]", write: Error},
	}
	cases := []struct {
		name    string
		message string
		want    string
	}{
		{name: "ordinary message", message: "ordinary message", want: "ordinary message"},
		{name: "carriage return", message: "first\rforged", want: `first\rforged`},
		{name: "line feed", message: "first\nforged", want: `first\nforged`},
		{name: "CRLF", message: "first\r\nforged", want: `first\r\nforged`},
		{name: "Unicode line separator", message: "first\u2028forged", want: `first\u2028forged`},
		{name: "Unicode paragraph separator", message: "first\u2029forged", want: `first\u2029forged`},
		{name: "terminal escape", message: "plain\x1b[31mred\x1b[0m", want: `plain\x1b[31mred\x1b[0m`},
		{name: "terminal CSI", message: "plain\u009b31mred", want: `plain\x9b31mred`},
	}
	for _, level := range levels {
		for _, test := range cases {
			t.Run(level.name+"/"+test.name, func(t *testing.T) {
				output.Reset()
				level.write("payload=%s", test.message)
				want := level.severity + " payload=" + test.want + "\n"
				if output.String() != want {
					t.Fatalf("output = %q, want %q", output.String(), want)
				}
				if strings.Count(output.String(), "\n") != 1 {
					t.Fatalf("physical line count = %d, want 1", strings.Count(output.String(), "\n"))
				}
			})
		}
	}
}

func TestLogFormattingRemainsReadable(t *testing.T) {
	var output bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	originalLevel := currentLevel
	log.SetOutput(&output)
	log.SetFlags(0)
	log.SetPrefix("")
	Init("debug")
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
		currentLevel = originalLevel
	})

	wrapped := fmt.Errorf("outer: %w", errors.New("inner"))
	Info("user %s made %d attempts: %v", "alice", 3, wrapped)
	const want = "[INFO] user alice made 3 attempts: outer: inner\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestFatalNeutralizesControlCharactersBeforeExit(t *testing.T) {
	if os.Getenv("WIKI_GO_LOGGER_FATAL_HELPER") == "1" {
		log.SetFlags(0)
		log.SetPrefix("")
		Fatal("fatal=%s", "first\r\nsecond\u2028third\x1b[31m")
		return
	}

	command := exec.Command(os.Args[0], "-test.run=^TestFatalNeutralizesControlCharactersBeforeExit$")
	command.Env = append(os.Environ(), "WIKI_GO_LOGGER_FATAL_HELPER=1")
	output, err := command.CombinedOutput()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 1 {
		t.Fatalf("Fatal exit error = %v, want status 1; output: %q", err, output)
	}
	const want = "[ERROR] fatal=first\\r\\nsecond\\u2028third\\x1b[31m\n"
	if string(output) != want {
		t.Fatalf("Fatal output = %q, want %q", output, want)
	}
	if bytes.Count(output, []byte{'\n'}) != 1 {
		t.Fatalf("Fatal physical line count = %d, want 1", bytes.Count(output, []byte{'\n'}))
	}
}
