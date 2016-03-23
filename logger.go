// Copyright (c) 2016 Ventu.io, Oleg Sklyar, contributors
// The use of this source code is governed by a MIT style license found in the LICENSE file

package slog

import (
	"errors"
	"fmt"
	"github.com/ventu-io/slf"
	"path"
	"runtime"
	"sync"
	"time"
)

const (
	// TraceField defined the key for the field to store trace duration.
	TraceField = "trace"

	// CallerField defines the key for the caller information.
	CallerField = "caller"

	// ErrorField can be used by handlers to represent the error in the data field collection.
	ErrorField = "error"

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
	provider *factory
}

// logger represents a logger in the context. It is created from the rootlogger by copying its
// fields. "Fields" access is not directly synchronised because fields are written on copy only,
// however it is synchronised indirectly to guarantee timestamp for tracing.
type logger struct {
	*rootlogger
	sync.Mutex
	fields    map[string]interface{}
	caller    slf.CallerInfo
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

// WithCaller implements the Logger interface.
func (log *logger) WithCaller(caller slf.CallerInfo) slf.StructuredLogger {
	res := log.copy()
	res.caller = caller
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
	return log.log(level, message)
}

// Trace implements the Logger interface.
func (log *logger) Trace(err *error) {
	log.Lock()
	defer log.Unlock()
	if log.lasttouch != epoch && log.lastlevel >= log.rootlogger.minlevel {
		var entry *entry
		if err != nil {
			entry = log.entry(log.lastlevel, traceMessage, 2, *err)
		} else {
			entry = log.entry(log.lastlevel, traceMessage, 2, nil)
		}
		entry.fields[TraceField] = time.Now().Sub(log.lasttouch)
		log.handle(entry)
	}
	// reset in any case
	log.lasttouch = epoch
}

// Debug implements the Logger interface.
func (log *logger) Debug(message string) slf.Tracer {
	return log.log(slf.LevelDebug, message)
}

// Debugf implements the Logger interface.
func (log *logger) Debugf(format string, args ...interface{}) slf.Tracer {
	return log.logf(format, slf.LevelDebug, args...)
}

// Info implements the Logger interface.
func (log *logger) Info(message string) slf.Tracer {
	return log.log(slf.LevelInfo, message)
}

// Infof implements the Logger interface.
func (log *logger) Infof(format string, args ...interface{}) slf.Tracer {
	return log.logf(format, slf.LevelInfo, args...)
}

// Warn implements the Logger interface.
func (log *logger) Warn(message string) slf.Tracer {
	return log.log(slf.LevelWarn, message)
}

// Warnf implements the Logger interface.
func (log *logger) Warnf(format string, args ...interface{}) slf.Tracer {
	return log.logf(format, slf.LevelWarn, args...)
}

// Error implements the Logger interface.
func (log *logger) Error(message string) slf.Tracer {
	return log.log(slf.LevelError, message)
}

// Errorf implements the Logger interface.
func (log *logger) Errorf(format string, args ...interface{}) slf.Tracer {
	return log.logf(format, slf.LevelError, args...)
}

// Panic implements the Logger interface.
func (log *logger) Panic(message string) {
	log.log(slf.LevelPanic, message)
	panic(errors.New(message))
}

// Panicf implements the Logger interface.
func (log *logger) Panicf(format string, args ...interface{}) {
	log.logf(format, slf.LevelPanic, args...)
	panic(fmt.Errorf(format, args...))
}

// Log implements the Logger interface.
func (log *logger) log(level slf.Level, message string) slf.Tracer {
	if level < log.rootlogger.minlevel {
		return noop
	}
	return log.checkedlog(level, message)
}

func (log *logger) logf(format string, level slf.Level, args ...interface{}) slf.Tracer {
	if level < log.rootlogger.minlevel {
		return noop
	}
	message := fmt.Sprintf(format, args...)
	return log.checkedlog(level, message)
}

func (log *logger) checkedlog(level slf.Level, message string) slf.Tracer {
	log.Lock()
	defer log.Unlock()
	log.handle(log.entry(level, message, 4, log.err))
	log.lasttouch = time.Now()
	log.lastlevel = level
	return log
}

func (log *logger) copy() *logger {
	res := &logger{
		rootlogger: log.rootlogger,
		fields:     make(map[string]interface{}),
		caller:     log.caller,
	}
	for key, value := range log.fields {
		res.fields[key] = value
	}
	return res
}

func (log *logger) entry(level slf.Level, message string, skip int, err error) *entry {
	fields := make(map[string]interface{})
	for key, value := range log.fields {
		fields[key] = value
	}
	if log.caller == slf.CallerLong || log.caller == slf.CallerShort {
		if _, file, line, ok := runtime.Caller(skip); ok {
			if log.caller == slf.CallerShort {
				file = path.Base(file)
			}
			fields[CallerField] = fmt.Sprintf("%s:%d", file, line)
		}
	}
	return &entry{tm: time.Now(), level: level, message: message, err: err, fields: fields}
}

func (log *logger) handle(entry *entry) {
	// unsafe wrt changing handlers (those should be initialized up front)
	for _, handler := range log.rootlogger.provider.handlers {
		go handler.Handle(entry)
	}
}
