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

// Logger represents an interface wrapper for the structured logger implementation of SLF.
type Logger interface {
	slf.Logger
	SetLevel(level slf.Level, contexts ...string)
	AddEntryHandler(handler EntryHandler)
	SetEntryHandlers(handlers ...EntryHandler)
	Contexts() map[string]slf.StructuredLogger
}

// New constructs a new logger conforming with SLF.
func New() Logger {
	res := &factory{
		root: rootlogger{
			minlevel: slf.LevelInfo,
		},
		contexts: make(map[string]*logger),
	}
	res.root.provider = res
	return res
}

// factory implements the slog.Logger interface.
type factory struct {
	sync.Mutex
	root     rootlogger
	contexts map[string]*logger
	handlers []EntryHandler
}

// WithContext delivers a logger for the given context (reusing loggers for the same context).
func (log *factory) WithContext(context string) slf.StructuredLogger {
	log.Lock()
	defer log.Unlock()
	return log.withContext(context)
}

func (log *factory) withContext(context string) *logger {
	ctx, ok := log.contexts[context]
	if ok {
		return ctx
	}
	fields := make(map[string]interface{})
	fields[ContextField] = context
	ctx = &logger{
		rootlogger: &rootlogger{minlevel: log.root.minlevel, provider: log.root.provider},
		fields:     fields,
	}
	log.contexts[context] = ctx
	return ctx
}

// SetLevel sets the logging slf.Level to given contexts, all loggers if no context given, or the root
// logger when context defined as "root".
func (log *factory) SetLevel(level slf.Level, contexts ...string) {
	log.Lock()
	defer log.Unlock()
	if len(contexts) == 0 {
		log.root.minlevel = level
		for _, logger := range log.contexts {
			logger.rootlogger.minlevel = level
		}
	} else {
		for _, context := range contexts {
			if strings.ToLower(context) != rootLevelKey {
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
func (log *factory) AddEntryHandler(handler EntryHandler) {
	log.handlers = append(log.handlers, handler)
}

// SetEntryHandlers overwrites existing entry handlers with a new set.
func (log *factory) SetEntryHandlers(handlers ...EntryHandler) {
	log.handlers = append([]EntryHandler{}, handlers...)
}

// Contexts returns all defined root logging contexts.
func (log *factory) Contexts() map[string]slf.StructuredLogger {
	res := make(map[string]slf.StructuredLogger)
	log.Lock()
	defer log.Unlock()
	for key, val := range log.contexts {
		res[key] = val
	}
	return res
}
