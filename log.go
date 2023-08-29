package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type Logger interface {
	Local(opts ...Op) Logger
	SetLevel(level Level) error
	Sync()
	Infof(format string, v ...any)
	Info(message string)
	Errorf(format string, v ...any)
	Error(message string)
	Warningf(format string, v ...any)
	Warning(message string)
	Debugf(format string, v ...any)
	Debug(message string)
	Fatalf(format string, v ...any)
	Fatal(message string)
}

type Level string

const (
	Fatal   Level = "FATAL"
	Error   Level = "ERROR"
	Warning Level = "WARN"
	Info    Level = "INFO"
	Debug   Level = "DEBUG"
)

var levels = map[Level]int{
	Debug:   0,
	Info:    1,
	Warning: 2,
	Error:   3,
	Fatal:   4,
}

const (
	DefaultLevel      = Info
	DefaultName       = "<...>"
	DefaultFormat     = FormatText
	DefaultWriteMode  = ModeBlocking
	DefaultDateFormat = "2006/01/02 15:04:05"
)

type globalProps struct {
	writers    []writer
	mode       WriteMode
	dateFormat string
	sync.WaitGroup
}

type localProps struct {
	writers []writer
	name    string
	level   int
}

type writer struct {
	writer io.Writer
	format Format
}

type Op func(*globalProps, *localProps)

func WithName(n string) Op {
	return func(gp *globalProps, lp *localProps) {
		if lp == nil {
			return
		}
		lp.name = n
	}
}

func WithLevel(l Level) Op {
	return func(gp *globalProps, lp *localProps) {
		if lp == nil {
			return
		}
		if v, err := parseLevel(l); err == nil {
			lp.level = v
		}
	}
}

func WithMode(m WriteMode) Op {
	return func(gp *globalProps, lp *localProps) {
		if gp == nil {
			return
		}
		gp.mode = m
	}
}

func WithDateFormat(f string) Op {
	return func(gp *globalProps, lp *localProps) {
		if gp == nil {
			return
		}
		gp.dateFormat = f
	}
}

func WithWriter(w io.Writer, f Format) Op {
	return func(gp *globalProps, lp *localProps) {
		if gp != nil {
			gp.writers = append(gp.writers, writer{w, f})
		} else if lp != nil {
			lp.writers = append(lp.writers, writer{w, f})
		}
	}
}

type log struct {
	global *globalProps
	local  *localProps
}

func New(opts ...Op) Logger {
	gp := &globalProps{
		dateFormat: DefaultDateFormat,
		mode:       DefaultWriteMode,
	}

	lp := &localProps{
		name:  DefaultName,
		level: levels[DefaultLevel],
	}

	for _, opt := range opts {
		opt(gp, lp)
	}

	if len(gp.writers) == 0 {
		gp.writers = append(gp.writers, writer{
			writer: os.Stdout,
			format: FormatTextColor,
		})
	}

	return &log{
		global: gp,
		local:  lp,
	}
}

func (l *log) Local(opts ...Op) Logger {
	lp := *l.local

	for _, opt := range opts {
		opt(nil, &lp)
	}

	return &log{
		global: l.global,
		local:  &lp,
	}
}

func (l *log) SetLevel(level Level) error {
	v, err := parseLevel(level)
	if err != nil {
		return err
	}

	l.local.level = v
	return nil
}

func (l *log) Sync() {
	l.global.Wait()
}

func parseLevel(level Level) (int, error) {
	if v, ok := levels[Level(strings.ToUpper(string(level)))]; ok {
		return v, nil
	}

	return 0, fmt.Errorf("incorrect log level type: %s", string(level))
}
