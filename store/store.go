package store

import (
	"io"
)

type Address struct {
	Score    string // The hash of the blob.
	Location string // Where a blob is stored. The format of this string is implementation specific
	Size     int64
	Offset   int64
}

// Store is an interface that represents a content addressable store for packrat
//
// Put will copy bytes from reader r to the storage system. Put will calculate the hash of the blob before uploading. It will determine whether or not to actually move the bytes.
//
// Get looks up the full address of bytes in the store and writes them to w.
//
// GetAddress writes bytes from the store for Address a to w.
//
// Describe looks up a blob in the store's index at id and returns the Address of the blob.
type Store interface {
	Put(r io.Reader) (Address, error)
	Get(score string, w io.Writer) error
	GetAddress(a Address, w io.Writer) error
	Describe(score string) (Address, error)
}
