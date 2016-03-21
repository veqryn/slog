// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog

import (
	"github.com/ventu-io/slf"
	"strings"
	"sync"
)

const (
	contextKey = "context"
)

type Logger struct {
	sync.Mutex
	root     rootlogger
	contexts map[string]*logger
	handlers []EntryHandler
}

func New() *Logger {
	res := &Logger{
		root: rootlogger{
			minlevel: slf.LevelInfo,
		},
		contexts: make(map[string]*logger),
	}
	res.root.provider = res
	return res
}

// WithContext delivers a logger for the given context (reusing loggers for the same context).
func (log *Logger) WithContext(context string) slf.StructuredLogger {
	log.Lock()
	defer log.Unlock()
	return log.withContext(context)
}

func (log *Logger) withContext(context string) *logger {
	ctx, ok := log.contexts[context]
	if ok {
		return ctx
	}
	fields := make(map[string]interface{})
	fields[contextKey] = context
	ctx = &logger{
		rootlogger: &rootlogger{minlevel: log.root.minlevel, provider: log.root.provider},
		fields:     fields,
	}
	log.contexts[context] = ctx
	return ctx
}

// SetLevel sets the logging slf.Level to given contexts, all loggers if no context given, or the root
// logger when context defined as "root".
func (log *Logger) SetLevel(level slf.Level, contexts ...string) {
	log.Lock()
	defer log.Unlock()
	if len(contexts) == 0 {
		for _, logger := range log.contexts {
			logger.rootlogger.minlevel = level
		}
	} else {
		for _, context := range contexts {
			if strings.ToLower(context) != "root" {
				log.withContext(context).rootlogger.minlevel = level
			} else {
				log.root.minlevel = level
			}
		}
	}
}

// AddEntryHandler adds a handler for log entries that are logged at or above the set
// log slf.Level. Unsafe if called while logging is already being performed, thus should be called
// at the initialisation time only.
func (log *Logger) AddEntryHandler(handler EntryHandler) {
	log.handlers = append(log.handlers, handler)
}
