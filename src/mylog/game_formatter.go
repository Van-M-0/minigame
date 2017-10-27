package mylog

import (
"bytes"
"fmt"
"sort"
"strings"
"sync"
"time"
)


func init() {
	baseTimestamp = time.Now()
}

type GameFormatter struct {

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// Enable logging the full timestamp when a TTY is attached instead of just
	// the time passed since beginning of execution.
	FullTimestamp bool


	// The fields are sorted by default for a consistent output. For applications
	// that log extremely frequently and don't use the JSON formatter this may not
	// be desired.
	DisableSorting bool

	// QuoteEmptyFields will wrap empty fields in quotes if true
	QuoteEmptyFields bool

	// QuoteCharacter can be set to the override the default quoting character "
	// with something else. For example: ', or `.
	QuoteCharacter string

	// Whether the logger's out is to a terminal
	isTerminal bool

	sync.Once
}

func (f *GameFormatter) init(entry *Entry) {
	if len(f.QuoteCharacter) == 0 {
		f.QuoteCharacter = "\""
	}
	if entry.Logger != nil {
		f.isTerminal = IsTerminal(entry.Logger.Out)
	}
}

func (f *GameFormatter) Format(entry *Entry) ([]byte, error) {
	var b *bytes.Buffer
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}

	if !f.DisableSorting {
		sort.Strings(keys)
	}
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.Do(func() { f.init(entry) })

	timestampFormat := "02 15:04:05.000"

	f.directWrite(b, entry, keys, timestampFormat)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *GameFormatter) directWrite(b *bytes.Buffer, entry *Entry, keys []string, timestampFormat string) {


	levelText := strings.ToUpper(entry.Level.String())[0:4]

	fmt.Fprintf(b, "%v [%s]	%-44s", entry.Time.Format(timestampFormat), levelText, entry.Message)
}

func (f *GameFormatter) needsQuoting(text string) bool {
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.') {
			return true
		}
	}
	return false
}

func (f *GameFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
	b.WriteByte(' ')
}

func (f *GameFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	switch value := value.(type) {
	case string:
		if !f.needsQuoting(value) {
			b.WriteString(value)
		} else {
			fmt.Fprintf(b, "%s%v%s", f.QuoteCharacter, value, f.QuoteCharacter)
		}
	case error:
		errmsg := value.Error()
		if !f.needsQuoting(errmsg) {
			b.WriteString(errmsg)
		} else {
			fmt.Fprintf(b, "%s%v%s", f.QuoteCharacter, errmsg, f.QuoteCharacter)
		}
	default:
		fmt.Fprint(b, value)
	}
}