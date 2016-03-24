// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog

import (
	"github.com/ventu-io/slf"
	"strings"
	"sync"
)

const (
	// ContextField defines the field to store context.
	ContextField = "context"
	rootLevelKey = "root"
)

// LogFactory represents an interface wrapper for the structured logger implementation of SLF.
type LogFactory interface {
	slf.LogFactory
	SetLevel(level slf.Level, contexts ...string)
	AddEntryHandler(handler EntryHandler)
	SetEntryHandlers(handlers ...EntryHandler)
	Contexts() map[string]slf.StructuredLogger
	SetConcurrent(conc bool)
}

// New constructs a new logger conforming with SLF.
func New() LogFactory {
	res := &logFactory{
		root: rootLogger{
			minlevel: slf.LevelInfo,
		},
		contexts:   make(map[string]*logger),
		concurrent: true,
	}
	res.root.factory = res
	return res
}

// factory implements the slog.Logger interface.
type logFactory struct {
	sync.Mutex
	root       rootLogger
	contexts   map[string]*logger
	handlers   []EntryHandler
	concurrent bool
}

// WithContext delivers a logger for the given context (reusing loggers for the same context).
func (lf *logFactory) WithContext(context string) slf.StructuredLogger {
	lf.Lock()
	defer lf.Unlock()
	return lf.withContext(context)
}

func (lf *logFactory) withContext(context string) *logger {
	ctx, ok := lf.contexts[context]
	if ok {
		return ctx
	}
	fields := make(map[string]interface{})
	fields[ContextField] = context
	ctx = &logger{
		rootLogger: &rootLogger{minlevel: lf.root.minlevel, factory: lf.root.factory},
		fields:     fields,
	}
	lf.contexts[context] = ctx
	return ctx
}

// SetLevel sets the logging slf.Level to given contexts, all loggers if no context given, or the root
// logger when context defined as "root".
func (lf *logFactory) SetLevel(level slf.Level, contexts ...string) {
	lf.Lock()
	defer lf.Unlock()
	if len(contexts) == 0 {
		lf.root.minlevel = level
		for _, logger := range lf.contexts {
			logger.rootLogger.minlevel = level
		}
	} else {
		for _, context := range contexts {
			if strings.ToLower(context) != rootLevelKey {
				lf.withContext(context).rootLogger.minlevel = level
			} else {
				lf.root.minlevel = level
			}
		}
	}
}

// AddEntryHandler adds a handler for log entries that are logged at or above the set
// log slf.Level. Unsafe if called while logging is already being performed, thus should be called
// at the initialisation time only.
func (lf *logFactory) AddEntryHandler(handler EntryHandler) {
	lf.handlers = append(lf.handlers, handler)
}

// SetEntryHandlers overwrites existing entry handlers with a new set.
func (lf *logFactory) SetEntryHandlers(handlers ...EntryHandler) {
	lf.handlers = append([]EntryHandler{}, handlers...)
}

// Contexts returns all defined root logging contexts.
func (lf *logFactory) Contexts() map[string]slf.StructuredLogger {
	res := make(map[string]slf.StructuredLogger)
	lf.Lock()
	defer lf.Unlock()
	for key, val := range lf.contexts {
		res[key] = val
	}
	return res
}

// SetConcurrent toggles concurrency in handling log messages. If concurrent (default), the
// output sequence of entries is not guaranteed to be the same as log entries input sequence,
// although the timestamp will correspond the time of logging, not handling.
func (lf *logFactory) SetConcurrent(conc bool) {
	lf.concurrent = conc
}
