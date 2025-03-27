package main

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
}

type FileTransactionLogger struct {
	// Something, something, fields
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	// Something, something, logic
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	// Something, something, logic
}
