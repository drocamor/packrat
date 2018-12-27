package index

import "testing"

func testAddGetExists(idx Index, t *testing.T) {
	id := "foo"
	exists := idx.Exists(id)
	if exists != false {
		t.Errorf("idx.Exists should have returned false on a non existent entry id")
	}

	_, err := idx.Get(id)
	if err == nil {
		t.Errorf("idx.Get should have returned an error on a non existent entry id")
	}

	e := Entry{Id: id}

	err = idx.Add(e)
	if err != nil {
		t.Errorf("idx.Add should not have returned an error, but got %v", err)
	}

	exists = idx.Exists(id)
	if exists == false {
		t.Errorf("idx.Exists should have returned true on an existent entry id")
	}

	_, err = idx.Get(id)
	if err != nil {
		t.Errorf("idx.Get should have not have returned an error on a non existent entry id, got %v", err)
	}

}

func testAliasGetAliasUnAlias(idx Index, t *testing.T) {
	id := "foraliasing"
	alias := "bar"
	e := Entry{Id: id}

	err := idx.Add(e)
	if err != nil {
		t.Errorf("Could not add entry to index: %v", err)
	}

	_, err = idx.GetAlias(alias)
	if err == nil {
		t.Errorf("idx.GetAlias should have failed on a non existent alias")
	}

	err = idx.Alias(alias, id)
	if err != nil {
		t.Errorf("idx.Alias should not have errored when adding an alias to a valid entry. Got error: %v", err)
	}

	anotherId := "anotheridforaliasing"
	anotherEntry := Entry{Id: anotherId}
	err = idx.Add(anotherEntry)
	if err != nil {
		t.Errorf("idx.Add failed: %v", err)
	}
	err = idx.Alias(alias, anotherId)
	if err == nil {
		t.Errorf("idx.Alias should have errored when adding an alias that already exists")
	}

	got, err := idx.GetAlias(alias)
	if err != nil {
		t.Errorf("idx.GetAlias should not have errored getting an valid entry with a valid index. Got error: %v", err)
	}

	if got.Id != id {
		t.Errorf("Entry id froh idx.GetAlias mismatched. Expected %q, got %q", id, got.Id)
	}

	err = idx.UnAlias(alias)
	if err != nil {
		t.Errorf("idx.UnAlias should have returned nil, got: %v", err)
	}

	// Adding an alias for a non existent item
	err = idx.Alias(alias, "bar")
	if err == nil {
		t.Errorf("idx.Alias should have errored when adding an alias to a non existent item")
	}

}

func testRelateRelationsUnrelate(idx Index, t *testing.T) {
	ids := []string{"a", "b", "c"}
	for _, id := range ids {
		e := Entry{Id: id}
		err := idx.Add(e)
		if err != nil {
			t.Errorf("Error adding to index: %v", err)
		}
	}

	relations := idx.Relations("a")
	if len(relations) != 0 {
		t.Errorf("foo should not have relations now, but it had %d", len(relations))
	}

	for _, id := range ids[1:] {
		err := idx.Relate(ids[0], id)
		if err != nil {
			t.Errorf("Unable to add relation. Error: %v", err)
		}
	}

	relations = idx.Relations("a")

	if len(relations) != len(ids[1:]) {
		t.Errorf("Number of relations doesn't match expected. Expected %d, got %d", len(ids[1:]), len(relations))
	}

	err := idx.UnRelate("a", ids[2])
	if err != nil {
		t.Errorf("Could not unrelate item. Error: %v", err)
	}

	relations = idx.Relations("a")
	if len(relations) != 1 {
		t.Errorf("Unrelating did not remove relations")
	}

	err = idx.Relate("a", "quux")
	if err == nil {
		t.Errorf("idx.Relate should have errored when attempting to relate something to a non existent entry")
	}

	err = idx.UnRelate("a", ids[1])
	if err != nil {
		t.Errorf("Could not unrelate item. Error: %v", err)
	}
}
