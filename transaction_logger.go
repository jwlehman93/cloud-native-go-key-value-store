package main

import (
	"fmt"
	"os"
)

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
}

type FileTransactionLogger struct {
	events       chan<- Event
	errors       <-chan error // Read-only channel for receiving errors
	lastSequence uint64       // The last used event sequence number
	file         *os.File     // The location of the transaction log
}

func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	return &FileTransactionLogger{file: file}, nil
}

func (l *FileTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events
	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		for e := range events {
			l.lastSequence++
			_, err := fmt.Fprintf(l.file, "%d\t%d\t%s\t%s\n", l.lastSequence, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

type Event struct {
	Sequence  uint64    // A unique record ID
	EventType EventType // The action taken
	Key       string
	Value     string
}

type EventType byte

const (
	_                     = iota
	EventDelete EventType = iota
	EventPut
)
