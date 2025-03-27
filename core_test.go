package main

import (
	"errors"
	"testing"
)

func TestPut(t *testing.T) {
	const key = "test-key"
	const value = "test-value"

	defer delete(store.m, key)

	var contains bool
	var val any

	_, contains = store.m[key]
	if contains {
		t.Error("key/value already exists")
	}

	err := Put(key, value)
	if err != nil {
		t.Error(err)
	}

	val, contains = store.m[key]

	if !contains {
		t.Error("create failed")
	}

	if val != value {
		t.Error("val/value mismatch")
	}
}

func TestGet(t *testing.T) {
	const key = "test-key"
	const value = "test-value"
	defer delete(store.m, key)

	var val any
	var err error

	val, err = Get(key)
	if err == nil {
		t.Error("Expected an error")
	}

	if !errors.Is(err, ErrorNoSuchKey) {
		t.Error("unexpected error:", err)
	}

	store.m[key] = value

	val, err = Get(key)
	if err != nil {
		t.Error("unexpected err:", err)
	}

	if val != value {
		t.Error("val/value mismatch")
	}
}

func TestDelete(t *testing.T) {
	const key = "test-key"
	const value = "test-value"
	defer delete(store.m, key)

	var contains bool

	store.m[key] = value

	_, contains = store.m[key]

	if !contains {
		t.Error("key/value does not exist")
	}

	err := Delete(key)
	if err != nil {
		t.Error("unexpected err:", err)
	}
	_, contains = store.m[key]
	if contains {
		t.Error("Delete failed")
	}

}
