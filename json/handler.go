// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

// Package json provides a JSON log entry handler formatting JSON into the given Writer.
package json

import (
	"encoding/json"
	"fmt"
	"github.com/ventu-io/slf"
	"github.com/ventu-io/slog"
	"io"
)

// TODO add handing log in batches (on size or time interval)

const (
	// StandardTimeFormat represents the time format used in the handler by default.
	StandardTimeFormat = "2006-01-02T15:04:05.0000"
)

// Handler represents a JSON log entry handler formatting JSON into the given Writer.
type Handler struct {
	writer        io.Writer
	timeFormatStr string
}

// New constructs a JSON handler formatting JSON into the given Writer.
func New(w io.Writer) *Handler {
	return &Handler{
		writer:        w,
		timeFormatStr: StandardTimeFormat,
	}
}

// SetTimeFormat defines the formatting of time used for output into JSON.
func (h *Handler) SetTimeFormat(f string) {
	h.timeFormatStr = f
}

type jsonentry struct {
	Timestamp string                  `json:"timestamp"`
	Level     slf.Level               `json:"level"`
	Message   string                  `json:"message"`
	Error     *string                 `json:"error"`
	Fields    *map[string]interface{} `json:"fields"`
}

// Handle processes the log entry formatting JSON into the given Writer.
func (h *Handler) Handle(e slog.Entry) (err error) {
	je := &jsonentry{
		Timestamp: e.Time().Format(h.timeFormatStr),
		Level:     e.Level(),
		Message:   e.Message(),
	}
	if e.Error() != nil {
		errs := e.Error().Error()
		je.Error = &errs
	}
	efields := e.Fields()
	if len(efields) > 0 {
		je.Fields = &efields
	}

	s, err := json.Marshal(je)
	if err != nil {
		return err
	}
	n, err := h.writer.Write(s)
	if err != nil {
		return err
	}
	if n != len(s) {
		return fmt.Errorf("json.Handler: Wrote only %v bytes out of %v", n, len(s))
	}
	return nil
}
