package zlog

// modified version of https://github.com/nxtgo/zlog, released under
// public domain.

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

func (l Level) String() string {
	return [...]string{"DBG", "INF", "WRN", "ERR", "FTL"}[l]
}

type F map[string]any

var (
	ansiReset = "\u001b[0m"
	ansiBlack = "\u001b[30m"

	defaultLevelColor = map[Level]string{
		LevelDebug: "\u001b[37m",
		LevelInfo:  "\u001b[34m",
		LevelWarn:  "\u001b[33m",
		LevelError: "\u001b[31m",
		LevelFatal: "\u001b[35m",
	}
	defaultLevelColorBg = map[Level]string{
		LevelDebug: "\u001b[47m",
		LevelInfo:  "\u001b[44m",
		LevelWarn:  "\u001b[43m",
		LevelError: "\u001b[41m",
		LevelFatal: "\u001b[45m",
	}
)

type Logger struct {
	mu           sync.Mutex
	out          io.Writer
	level        Level
	timeStamp    bool
	timeFormat   string
	json         bool
	colors       bool
	caller       bool
	fields       F
	levelColor   map[Level]string
	levelColorBg map[Level]string
}

var Log = new()

func new() *Logger {
	return &Logger{
		out:          os.Stderr,
		level:        LevelDebug,
		timeStamp:    true,
		timeFormat:   time.RFC3339,
		colors:       isTerminal(os.Stderr),
		caller:       true,
		fields:       make(F),
		levelColor:   maps.Clone(defaultLevelColor),
		levelColorBg: maps.Clone(defaultLevelColorBg),
	}
}

func (l *Logger) SetLevelColor(level Level, fg string) {
	l.mu.Lock()
	l.levelColor[level] = fg
	l.mu.Unlock()
}
func (l *Logger) SetLevelBgColor(level Level, bg string) {
	l.mu.Lock()
	l.levelColorBg[level] = bg
	l.mu.Unlock()
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
	if f, ok := w.(*os.File); ok {
		l.colors = isTerminal(f)
	}
}
func (l *Logger) SetLevel(level Level)     { l.mu.Lock(); l.level = level; l.mu.Unlock() }
func (l *Logger) EnableTimestamps(on bool) { l.mu.Lock(); l.timeStamp = on; l.mu.Unlock() }
func (l *Logger) SetTimeFormat(tf string)  { l.mu.Lock(); l.timeFormat = tf; l.mu.Unlock() }
func (l *Logger) SetJSON(on bool)          { l.mu.Lock(); l.json = on; l.mu.Unlock() }
func (l *Logger) EnableColors(on bool)     { l.mu.Lock(); l.colors = on; l.mu.Unlock() }
func (l *Logger) ShowCaller(on bool)       { l.mu.Lock(); l.caller = on; l.mu.Unlock() }

func (l *Logger) WithFields(f F) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	newFields := make(F, len(l.fields)+len(f))
	maps.Copy(newFields, l.fields)
	maps.Copy(newFields, f)
	return &Logger{
		out: l.out, level: l.level,
		timeStamp: l.timeStamp, timeFormat: l.timeFormat,
		json: l.json, colors: l.colors, caller: l.caller,
		fields:       newFields,
		levelColor:   maps.Clone(l.levelColor),
		levelColorBg: maps.Clone(l.levelColorBg),
	}
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func (l *Logger) Log(level Level, msg string, extra F) {
	l.mu.Lock()
	out, jsonMode, colors, timeStamp, tf, callerOn := l.out, l.json, l.colors, l.timeStamp, l.timeFormat, l.caller
	base := make(F, len(l.fields))
	maps.Copy(base, l.fields)
	l.mu.Unlock()

	if level < l.level {
		return
	}
	maps.Copy(base, extra)

	var callerStr string
	if callerOn {
		if _, file, line, ok := runtime.Caller(3); ok {
			callerStr = fmt.Sprintf("%s:%d", shortFile(file), line)
		}
	}

	if jsonMode {
		entry := make(map[string]any, len(base)+4)
		entry["level"], entry["msg"] = level.String(), msg
		if timeStamp {
			entry["time"] = time.Now().Format(tf)
		}
		if callerStr != "" {
			entry["caller"] = callerStr
		}
		maps.Copy(entry, base)
		if b, err := json.Marshal(entry); err != nil {
			fmt.Fprintf(out, "json marshal error: %v\n", err)
			return
		} else {
			fmt.Fprintln(out, string(b))
		}
		if level == LevelFatal {
			os.Exit(1)
		}
		return
	}

	var b strings.Builder
	if colors {
		if c, ok := l.levelColor[level]; ok {
			b.WriteString(c)
		}
	}
	if timeStamp {
		b.WriteString(time.Now().Format(tf) + " ")
	}
	if colors {
		if c, ok := l.levelColorBg[level]; ok {
			b.WriteString(c + ansiBlack)
		}
	}

	b.WriteString(" " + level.String() + " ")
	if colors {
		if c, ok := l.levelColor[level]; ok {
			b.WriteString(ansiReset + c)
		}
	}
	b.WriteString(" " + msg)

	if len(base) > 0 {
		b.WriteString(" ")
		keys := make([]string, 0, len(base))
		for k := range base {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				b.WriteString(" ")
			}
			fmt.Fprintf(&b, "%s=%v", k, base[k])
		}
	}
	if callerStr != "" {
		fmt.Fprintf(&b, " (%s)", callerStr)
	}
	if colors {
		b.WriteString(ansiReset)
	}

	fmt.Fprintln(out, b.String())
	if level == LevelFatal {
		os.Exit(1)
	}
}

func shortFile(path string) string {
	parts := strings.Split(path, "/")
	if n := len(parts); n >= 2 {
		return strings.Join(parts[n-2:], "/")
	}
	return path
}

func (l *Logger) Debug(msg string) { l.Log(LevelDebug, msg, nil) }
func (l *Logger) Info(msg string)  { l.Log(LevelInfo, msg, nil) }
func (l *Logger) Warn(msg string)  { l.Log(LevelWarn, msg, nil) }
func (l *Logger) Error(msg string) { l.Log(LevelError, msg, nil) }
func (l *Logger) Fatal(msg string) { l.Log(LevelFatal, msg, nil) }

func (l *Logger) Debugf(f string, a ...any) { l.Log(LevelDebug, fmt.Sprintf(f, a...), nil) }
func (l *Logger) Infof(f string, a ...any)  { l.Log(LevelInfo, fmt.Sprintf(f, a...), nil) }
func (l *Logger) Warnf(f string, a ...any)  { l.Log(LevelWarn, fmt.Sprintf(f, a...), nil) }
func (l *Logger) Errorf(f string, a ...any) { l.Log(LevelError, fmt.Sprintf(f, a...), nil) }
func (l *Logger) Fatalf(f string, a ...any) { l.Log(LevelFatal, fmt.Sprintf(f, a...), nil) }

func (l *Logger) Debugw(msg string, f F, a ...any) { l.Log(LevelDebug, fmt.Sprintf(msg, a...), f) }
func (l *Logger) Infow(msg string, f F, a ...any)  { l.Log(LevelInfo, fmt.Sprintf(msg, a...), f) }
func (l *Logger) Warnw(msg string, f F, a ...any)  { l.Log(LevelWarn, fmt.Sprintf(msg, a...), f) }
func (l *Logger) Errorw(msg string, f F, a ...any) { l.Log(LevelError, fmt.Sprintf(msg, a...), f) }
func (l *Logger) Fatalw(msg string, f F, a ...any) { l.Log(LevelFatal, fmt.Sprintf(msg, a...), f) }

func SetOutput(w io.Writer)    { Log.SetOutput(w) }
func SetLevel(l Level)         { Log.SetLevel(l) }
func EnableTimestamps(on bool) { Log.EnableTimestamps(on) }
func SetTimeFormat(tf string)  { Log.SetTimeFormat(tf) }
func SetJSON(on bool)          { Log.SetJSON(on) }
func EnableColors(on bool)     { Log.EnableColors(on) }
func ShowCaller(on bool)       { Log.ShowCaller(on) }
func WithFields(f F) *Logger   { return Log.WithFields(f) }
