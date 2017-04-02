package log

import (
	"fmt"
	"io"
	"log"
)

type Logger struct {
	debug   bool
	logger  *log.Logger
	entries chan string
}

func New(wr io.Writer, debug bool) *Logger {
	l := &Logger{
		logger:  log.New(wr, "", log.Ldate|log.Ltime|log.LUTC),
		entries: make(chan string),
	}
	go l.write()
	return l
}

func (l *Logger) Debug(s string) {
	if !l.debug {
		return
	}
	go func() {
		l.entries <- s
	}()
}

func (l *Logger) Info(s string) {
	go func() {
		l.entries <- s
	}()
}

func (l *Logger) Err(err error) {
	go func() {
		l.entries <- err.Error()
	}()
}

func (l *Logger) write() {
	for s := range l.entries {
		l.logger.Println(s)
		fmt.Println(s)
	}
}
