package index

import "testing"

func TestInMemAddGetExists(t *testing.T) {
	idx := NewInMemoryIndex()

	testAddGetExists(&idx, t)
}

func TestInMemAliasGetAliasUnAlias(t *testing.T) {
	idx := NewInMemoryIndex()
	testAliasGetAliasUnAlias(&idx, t)
}

func TestInMemRelateRelationsUnrelate(t *testing.T) {
	idx := NewInMemoryIndex()
	testRelateRelationsUnrelate(&idx, t)
}
