// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog

import (
	"errors"
	"fmt"
	"github.com/ventu-io/slf"
	"sync"
	"time"
)

const (
	traceMessage = "trace"
)

var (
	noop  = &slf.Noop{}
	epoch = time.Time{}
)

// rootlogger represents a root logger for a context, all other loggers in the same context
// (with different fields) contain this one to identify the log level and entry handlers.
type rootlogger struct {
	minlevel slf.Level
	provider *Logger
}

// logger represents a logger in the context. It is created from the rootlogger by copying its
// fields. "Fields" access is not directly synchronised because fields are written on copy only,
// however it is synchronised indirectly to guarantee timestamp for tracing.
type logger struct {
	*rootlogger
	sync.Mutex
	fields    map[string]interface{}
	err       error
	lasttouch time.Time
	lastlevel slf.Level
}

// WithField implements the Logger interface.
func (log *logger) WithField(key string, value interface{}) slf.StructuredLogger {
	res := log.copy()
	res.fields[key] = value
	return res
}

// WithFields implements the Logger interface.
func (log *logger) WithFields(fields slf.Fields) slf.StructuredLogger {
	res := log.copy()
	for k, v := range fields {
		res.fields[k] = v
	}
	return res
}

// WithError implements the Logger interface.
func (log *logger) WithError(err error) slf.BasicLogger {
	res := log.copy()
	res.err = err
	return res
}

// Log implements the Logger interface.
func (log *logger) Log(level slf.Level, message string) slf.Tracer {
	if level < log.rootlogger.minlevel {
		return noop
	}
	log.Lock()
	defer log.Unlock()
	log.handle(level, message, log.err)
	log.lasttouch = time.Now()
	log.lastlevel = level
	return log
}

// Logf implements the Logger interface.
func (log *logger) Logf(level slf.Level, format string, args ...interface{}) slf.Tracer {
	if level < log.rootlogger.minlevel {
		return noop
	}
	message := fmt.Sprintf(format, args...)
	return log.Log(level, message)
}

// Trace implements the Logger interface.
func (log *logger) Trace(err *error) {
	log.Lock()
	defer log.Unlock()
	if log.lasttouch != epoch {
		if log.lastlevel >= log.rootlogger.minlevel {
			log.handle(log.lastlevel, traceMessage, *err)
		}
		// reset in any case
		log.lasttouch = epoch
	}
}

// Debug implements the Logger interface.
func (log *logger) Debug(message string) slf.Tracer {
	return log.Log(slf.LevelDebug, message)
}

// Debugf implements the Logger interface.
func (log *logger) Debugf(format string, args ...interface{}) slf.Tracer {
	return log.Logf(slf.LevelDebug, format, args...)
}

// Info implements the Logger interface.
func (log *logger) Info(message string) slf.Tracer {
	return log.Log(slf.LevelInfo, message)
}

// Infof implements the Logger interface.
func (log *logger) Infof(format string, args ...interface{}) slf.Tracer {
	return log.Logf(slf.LevelInfo, format, args...)
}

// Warn implements the Logger interface.
func (log *logger) Warn(message string) slf.Tracer {
	return log.Log(slf.LevelWarn, message)
}

// Warnf implements the Logger interface.
func (log *logger) Warnf(format string, args ...interface{}) slf.Tracer {
	return log.Logf(slf.LevelWarn, format, args...)
}

// Error implements the Logger interface.
func (log *logger) Error(message string) slf.Tracer {
	return log.Log(slf.LevelError, message)
}

// Errorf implements the Logger interface.
func (log *logger) Errorf(format string, args ...interface{}) slf.Tracer {
	return log.Logf(slf.LevelError, format, args...)
}

// Panic implements the Logger interface.
func (log *logger) Panic(message string) {
	log.Log(slf.LevelPanic, message)
	panic(errors.New(message))
}

// Panicf implements the Logger interface.
func (log *logger) Panicf(format string, args ...interface{}) {
	log.Logf(slf.LevelPanic, format, args...)
	panic(fmt.Errorf(format, args...))
}

func (log *logger) copy() *logger {
	res := &logger{
		rootlogger: log.rootlogger,
		fields:     make(map[string]interface{}),
	}
	for key, value := range log.fields {
		res.fields[key] = value
	}
	return res
}

func (log *logger) handle(level slf.Level, message string, err error) {
	entry := &entry{
		level:   level,
		message: message,
		err:     err,
		fields:  make(map[string]interface{}),
	}
	for key, value := range log.fields {
		entry.fields[key] = value
	}
	// unsafe wrt changing handlers (those should be initialized up front)
	for _, handler := range log.rootlogger.provider.handlers {
		go handler.Handle(entry)
	}
}
