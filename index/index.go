package index

import (
	"github.com/drocamor/packrat/store"
	"time"
)

type Entry struct {
	Id           string    // Concatenation of the item's timestamp and some random junk
	Name         string    `json:",omitempty"` // Name of the item. Not required.
	Timestamp    time.Time // When this item happened
	Importance   int       // Importance is an arbitrary number that lets you filter out things that are not important
	Type         string    `json:",omitempty"` // What kind of thing this is, used for like thumbnailing, etc
	Gridsquare   string    `json:",omitempty"` // maidenhead grid square
	Group        string    // The group that this belongs to. there is probably one group per installation.
	GridsquareId string    `json:",omitempty"` // concatenation of gridsquare and Id

	Addresses map[string]store.Address // A map of where the data is stored. Typically there is an original and an thumbnail
}

type Index interface {
	Add(entry Entry) error                // Add an item to the index.
	Get(id string) (Entry, error)         // Return a full entry from the index
	Exists(id string) bool                // Tell me if something is in the index or not
	Alias(alias, id string) error         // Adds an Alias to an entry
	GetAlias(alias string) (Entry, error) // Gets the entry for an alias
	UnAlias(alias string) error           // Removes an Alias to an entry
	Relate(a, b string) error             // Relates one ID to another ID
	UnRelate(a, b string) error           // deletes a relation
	Relations(id string) []string         // returns the relations for an entry
	// TODO Query
}
