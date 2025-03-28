package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresTransactionLogger struct {
	events chan<- Event // Write-only channel for sending events
	errors <-chan error // Read-only channel for receiving errors
	db     *sql.DB      // The database access interface
}

type PostgresDBParams struct {
	dbName   string
	host     string
	user     string
	password string
}

func NewPostgresTransactionLoggerWithConfig(config PostgresDBParams) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s",
		config.host, config.dbName, config.user, config.password)
	return NewPostgresTransactionLogger(connStr)
}

func NewPostgresTransactionLogger(connStr string) (TransactionLogger,
	error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	err = db.Ping() // Test the database connection
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	logger := &PostgresTransactionLogger{db: db}

	exists, err := logger.verifyTableExists()
	if err != nil {
		return nil, fmt.Errorf("failed to verify table exists: %w", err)
	}
	if !exists {
		if err = logger.createTable(); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

	return logger, nil
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events
	errors := make(chan error, 1)
	l.errors = errors
	insertEventQuery := `INSERT INTO transactions (event_type, key, value) VALUES ($1, $2, $3);`
	go func() {
		for e := range events {
			_, err := l.db.Exec(insertEventQuery, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvents := make(chan Event)
	outErrors := make(chan error, 1)
	go func() {
		defer close(outEvents)
		defer close(outErrors)
		selectQuery := "SELECT (sequence, event_type, key, value) FROM transactions ORDER BY sequence"
		rows, err := l.db.Query(selectQuery)
		if err != nil {
			outErrors <- err
			return
		}
		defer rows.Close()

		e := Event{}
		for rows.Next() {
			err := rows.Scan(&e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outErrors <- err
				return
			}
			outEvents <- e
		}
		err = rows.Err()
		if err != nil {
			outErrors <- err
			return
		}
	}()

	return outEvents, outErrors
}

func (l *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	const table = "transactions"
	rows, err := l.db.Query("SELECT to_regclass($1)", table)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var result string
	for rows.Next() && result != table {
		rows.Scan(&result)
	}

	return result == table, rows.Err()
}

func (l *PostgresTransactionLogger) createTable() error {
	createTableQuery := `CREATE TABLE IF NOT EXISTS transactions (
							sequence SERIAL PRIMARY KEY,
							event_type INT,
							key TEXT,
							value TEXT
						);`
	_, err := l.db.Exec(createTableQuery)
	if err != nil {
		return err
	}
	return nil
}

func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}
