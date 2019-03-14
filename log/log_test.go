package log

import (
	"fmt"
	"testing"
)

type mockLogger struct {
	level  Level
	logged string
	fatal  bool

	expected      string
	fatalExpected bool
}

func newMockLogger(level Level, expected string, fatal bool) *mockLogger {
	return &mockLogger{
		level:         level,
		expected:      expected,
		fatalExpected: fatal,
	}
}

func (l *mockLogger) Printf(format string, args ...interface{}) {
	l.logged += fmt.Sprintf(format, args...)
}

func (l *mockLogger) Print(args ...interface{}) {
	l.logged += fmt.Sprint(args...)
}

func (l *mockLogger) Fatal(args ...interface{}) {
	l.fatal = true
}

func (l *mockLogger) Fatalf(format string, args ...interface{}) {
	l.fatal = true
}

func (l *mockLogger) verify(t *testing.T, i int) {
	t.Helper()

	if l.logged != l.expected {
		t.Errorf("%.4d: expected %q, got %q", i, l.expected, l.logged)
	}
	if l.fatal != l.fatalExpected {
		t.Errorf("%.4d: expected fatal %v, got %v", i, l.fatalExpected, l.fatal)
	}
}

func testLevel(t *testing.T, ml *mockLogger, level Level) {
	SetLevel(level)

	l := New(ml, ml.level)
	l.Print("log line")
	l.Fatal("fatal line")
}

func TestLevel(t *testing.T) {
	for i, v := range []struct {
		mockLevel    Level
		expected     string
		fatal        bool
		currentLevel Level
	}{
		{DebugLevel, "log line", true, DebugLevel},
		{DebugLevel, "", true, InfoLevel},
		{DebugLevel, "", true, ErrorLevel},
		{DebugLevel, "", true, DisabledLevel},

		{InfoLevel, "log line", true, DebugLevel},
		{InfoLevel, "log line", true, InfoLevel},
		{InfoLevel, "", true, ErrorLevel},
		{InfoLevel, "", true, DisabledLevel},

		{ErrorLevel, "log line", true, DebugLevel},
		{ErrorLevel, "log line", true, InfoLevel},
		{ErrorLevel, "log line", true, ErrorLevel},
		{ErrorLevel, "", true, DisabledLevel},
	} {
		ml := newMockLogger(v.mockLevel, v.expected, v.fatal)
		testLevel(t, ml, v.currentLevel)
		ml.verify(t, i)
	}
}
