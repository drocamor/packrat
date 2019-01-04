package index

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"testing"
)

func deleteEntry(idx *DynamoDBIndex, id string) error {
	// Remove the entry
	key := map[string]*dynamodb.AttributeValue{
		"Group": {
			S: aws.String(idx.group),
		},
		"Id": {
			S: aws.String(id),
		},
	}

	params := (&dynamodb.DeleteItemInput{}).
		SetTableName(idx.entriesTable()).
		SetKey(key)

	_, err := idx.ddb.DeleteItem(params)
	return err
}

func TestDdb(t *testing.T) {
	// cfg := (&aws.Config{}).WithRegion("us-west-2")
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-west-2")}))
	idx := NewDynamoDBIndex(sess, "noone", "testPR")

	t.Run("AddGetExists", func(t *testing.T) { testAddGetExists(idx, t) })
	t.Run("AliasGetAliasUnAlias", func(t *testing.T) { testAliasGetAliasUnAlias(idx, t) })
	t.Run("RelateRelationsUnrelate", func(t *testing.T) { testRelateRelationsUnrelate(idx, t) })
	for _, k := range []string{"foo", "baz", "foraliasing", "anotheridforaliasing", "a", "b", "c"} {
		err := deleteEntry(idx, k)
		if err != nil {
			t.Errorf("Could not delete entry: %v", err)
		}
	}
}
