// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package text

import (
	"fmt"
	"github.com/ventu-io/slf"
	"github.com/ventu-io/slog"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"
)

const (
	none   int = 0
	red        = 31
	green      = 32
	yellow     = 33
	blue       = 34
	gray       = 37
)

var start = time.Now()

const DefaultTemplate = "{{.Time}} \033[{{.Color}}m[{{.Level}}]\033[0m {{.Context}} - {{.Message}}: {{.Fields}}\n"

type Holder struct {
	Time    string
	Level   string
	Context string
	Message string
	Caller  string
	Fields  string
	Color   int
}

type TextHandler struct {
	sync.Mutex
	writer      io.Writer
	colors      map[slf.Level]int
	templateStr string
	template    *template.Template
}

func New() *TextHandler {
	res := &TextHandler{
		writer:      os.Stderr,
		colors:      make(map[slf.Level]int),
		templateStr: DefaultTemplate,
		// TODO handle error/panic
		template: template.Must(template.New("entry").Parse(DefaultTemplate)),
	}
	res.colors[slf.LevelDebug] = blue
	res.colors[slf.LevelInfo] = green
	res.colors[slf.LevelWarn] = yellow
	res.colors[slf.LevelError] = red
	res.colors[slf.LevelPanic] = red
	return res
}

func (th *TextHandler) SetColors(colors map[slf.Level]int) error {
	// TODO validate
	th.colors = colors
	return nil
}

func (th *TextHandler) SetTemplate(t string) error {
	var err error
	th.template, err = template.New("entry").Parse(t)
	if err != nil {
		return err
	}
	th.templateStr = t
	return nil
}

func (th *TextHandler) SetWriter(writer io.Writer) error {
	th.writer = writer
	return nil
}

func tostr(level slf.Level) string {
	switch level {
	case slf.LevelDebug:
		return "DEBUG"
	case slf.LevelWarn:
		return "WARN"
	case slf.LevelError:
		return "ERROR"
	case slf.LevelPanic:
		return "PANIC"
	default:
		return "INFO"
	}
}

func (th *TextHandler) Handle(entry slog.Entry) error {
	level := tostr(entry.Level())
	context, ok := entry.Fields()[slog.ContextField]
	if !ok {
		context = "<context ?>"
	}
	caller, ok := entry.Fields()[slog.CallerField]
	if !ok {
		caller = "<caller ?>"
	}
	fps := []fieldpair{}
	for key, value := range entry.Fields() {
		if key == slog.ContextField && strings.Contains(th.templateStr, "{{.Context}}") {
			continue
		}
		if key == slog.CallerField && strings.Contains(th.templateStr, "{{.Caller}}") {
			continue
		}
		fps = append(fps, fieldpair{key, value})
	}
	if entry.Error() != nil {
		fps = append(fps, fieldpair{slog.ErrorField, entry.Error().Error()})
	}
	sort.Sort(byName(fps))

	fields := []string{}
	for _, fp := range fps {
		fields = append(fields, fmt.Sprintf("%s=%v", fp.Name, fp.Value))
	}

	color, ok := th.colors[entry.Level()]
	if !ok {
		color = gray
	}

	h := &Holder{
		Time:    entry.Time().Format("2006-01-02 15:04:05.0000"),
		Level:   level,
		Context: fmt.Sprint(context),
		Message: entry.Message(),
		Caller:  fmt.Sprint(caller),
		Fields:  strings.Join(fields, "; "),
		Color:   color,
	}
	th.Lock()
	defer func() {
		th.Unlock()
		if r := recover(); r != nil {
			log.Print(r)
		}
	}()
	return th.template.Execute(th.writer, h)
}

type fieldpair struct {
	Name  string
	Value interface{}
}

type byName []fieldpair

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
