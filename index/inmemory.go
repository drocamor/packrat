package index

import (
	"fmt"
	"sync"
)

type InMemoryIndex struct {
	entries                               map[string]Entry
	aliases                               map[string]string
	relations                             map[string]map[string]struct{}
	entryMutex, aliasMutex, relationMutex sync.Mutex
}

func NewInMemoryIndex() InMemoryIndex {
	return InMemoryIndex{
		entries:   make(map[string]Entry),
		aliases:   make(map[string]string),
		relations: make(map[string]map[string]struct{}),
	}
}

func (i *InMemoryIndex) Add(entry Entry) error {
	i.entryMutex.Lock()
	defer i.entryMutex.Unlock()

	i.entries[entry.Id] = entry

	return nil
}

func (i *InMemoryIndex) Get(id string) (Entry, error) {
	i.entryMutex.Lock()
	defer i.entryMutex.Unlock()

	e, ok := i.entries[id]
	if !ok {
		return e, fmt.Errorf("Entry does not exist in index")
	}

	return e, nil
}

func (i *InMemoryIndex) Exists(id string) bool {
	i.entryMutex.Lock()
	defer i.entryMutex.Unlock()

	_, exists := i.entries[id]
	return exists
}

func (i *InMemoryIndex) Alias(alias, id string) error {
	i.aliasMutex.Lock()
	defer i.aliasMutex.Unlock()

	_, aliasExists := i.aliases[alias]
	if aliasExists {
		return fmt.Errorf("Alias already exists")
	}

	if i.Exists(id) == false {
		return fmt.Errorf("Entry does not exist in index")
	}

	i.aliases[alias] = id
	return nil

}

func (i *InMemoryIndex) GetAlias(alias string) (Entry, error) {
	i.aliasMutex.Lock()
	defer i.aliasMutex.Unlock()

	id, ok := i.aliases[alias]
	if !ok {
		return Entry{}, fmt.Errorf("Alias does not exist in index")
	}
	return i.Get(id)
}

func (i *InMemoryIndex) UnAlias(alias string) error {
	i.aliasMutex.Lock()
	defer i.aliasMutex.Unlock()

	delete(i.aliases, alias)
	return nil
}

func (i *InMemoryIndex) Relate(a, b string) error {
	i.relationMutex.Lock()
	defer i.relationMutex.Unlock()

	for _, id := range []string{a, b} {
		if !i.Exists(id) {
			return fmt.Errorf("%q does not exist in index", id)
		}
	}

	rel, ok := i.relations[a]
	if !ok {
		rel = make(map[string]struct{})
	}

	rel[b] = struct{}{}

	i.relations[a] = rel
	return nil
}

func (i *InMemoryIndex) UnRelate(a, b string) error {
	i.relationMutex.Lock()
	defer i.relationMutex.Unlock()

	delete(i.relations[a], b)
	return nil
}

func (i *InMemoryIndex) Relations(id string) []string {
	i.relationMutex.Lock()
	defer i.relationMutex.Unlock()

	relations := make([]string, 0)
	relationsMap, ok := i.relations[id]
	if !ok {
		return relations
	}

	for id, _ := range relationsMap {
		relations = append(relations, id)
	}

	return relations

}
