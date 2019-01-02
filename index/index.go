package index

import (
	"fmt"
	"time"
	"github.com/drocamor/packrat/store"
	"github.com/satori/go.uuid"

)

type Entry struct {
	Id         string // a string representation of the sha256 hash for this item
	Name       string `json:",omitempty"` // Name of the item. Not required.
	Timestamp  time.Time
	Importance int // Importance is an arbitrary number, that lets you filter out things that are not important

	Type       string `json:",omitempty"` // What kind of thing this is, used for like thumbnailing, etc
	Gridsquare string `json:",omitempty"` // maidenhead grid square

	UserId       string // A user ID, to make this multitenant
	TimestampId  string `json:",omitempty"` // concatenation of timestamp and Id
	GridsquareId string `json:",omitempty"` // concatenation of gridsquare and Id

	Object    store.Address
	Thumbnail store.Address
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

func NewEntry() Entry {
	u := uuid.Must(uuid.NewV4())
	return Entry{
		Id:        fmt.Sprintf("%s", u),
		Timestamp: time.Now(),
	}

}
