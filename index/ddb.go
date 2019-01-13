package index

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
)

/*
Entries: timestamp-score
- GSI Location: username, location-score

Alias: alias
Relations: id, otherid

*/

const (
	entriesTable      = "Entries"
	gridsquareIdIndex = "Group-Gridsquare"
	aliasesTable      = "Aliases"
	relationsTable    = "Relations"
)

type dynamoDBAlias struct {
	Group, Alias, Id string
}

// dynamoDBRelation is a entry in the relations table.
// A is always a concatenation of the group, "-", and the id of the entry with the relationships
// B is the Id of the entries that are related to A
type dynamoDBRelation struct {
	A, B string
}

type DynamoDBIndex struct {
	ddb                *dynamodb.DynamoDB
	group, tablePrefix string
}

func NewDynamoDBIndex(sess *session.Session, group string, tablePrefix string) *DynamoDBIndex {
	ddb := dynamodb.New(sess)
	return &DynamoDBIndex{
		ddb:         ddb,
		group:       group,
		tablePrefix: tablePrefix,
	}
}

func (i *DynamoDBIndex) entriesTable() string {
	return i.tablePrefix + entriesTable
}

func (i *DynamoDBIndex) aliasesTable() string {
	return i.tablePrefix + aliasesTable
}

func (i *DynamoDBIndex) relationsTable() string {
	return i.tablePrefix + relationsTable
}

func (i *DynamoDBIndex) Add(entry Entry) error {
	entry.Group = i.group
	entry.createIds()
	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		return err
	}

	params := (&dynamodb.PutItemInput{}).
		SetTableName(i.entriesTable()).
		SetConditionExpression("attribute_not_exists(Id)").
		SetItem(av)

	_, err = i.ddb.PutItem(params)

	// Return a special error if the entry is already in the index
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return ErrAlreadyExists
			}
		}
	}

	return err

}

func (i *DynamoDBIndex) Get(id string) (Entry, error) {
	var entry Entry

	key := map[string]*dynamodb.AttributeValue{
		"Group": {
			S: aws.String(i.group),
		},
		"Id": {
			S: aws.String(id),
		},
	}

	params := (&dynamodb.GetItemInput{}).
		SetTableName(i.entriesTable()).
		SetKey(key)

	resp, err := i.ddb.GetItem(params)
	if err != nil {
		return entry, err
	}
	if resp.Item == nil {
		return entry, fmt.Errorf("Entry does not exist in index")
	}

	err = dynamodbattribute.UnmarshalMap(resp.Item, &entry)

	return entry, err
}
func (i *DynamoDBIndex) Exists(id string) bool {
	// TODO do this more efficiently
	_, err := i.Get(id)
	if err != nil {
		return false
	}
	return true
}
func (i *DynamoDBIndex) Alias(alias, id string) error {
	// checks that entry exists
	if i.Exists(id) != true {
		return fmt.Errorf("Entry does not exist in index.")
	}

	// Creates an alias in alias table.
	a := dynamoDBAlias{
		Group: i.group,
		Alias: alias,
		Id:    id,
	}

	// Conditional put based on if alias exists
	av, err := dynamodbattribute.MarshalMap(a)
	if err != nil {
		return err
	}

	params := (&dynamodb.PutItemInput{}).
		SetTableName(i.aliasesTable()).
		SetConditionExpression("attribute_not_exists(Alias)").
		SetItem(av)

	_, err = i.ddb.PutItem(params)
	return err
}

func (i *DynamoDBIndex) GetAlias(alias string) (Entry, error) {
	var entry Entry

	// Get on the Alias table
	key := map[string]*dynamodb.AttributeValue{
		"Group": {
			S: aws.String(i.group),
		},
		"Alias": {
			S: aws.String(alias),
		},
	}

	params := (&dynamodb.GetItemInput{}).
		SetTableName(i.aliasesTable()).
		SetKey(key)

	resp, err := i.ddb.GetItem(params)
	if err != nil {
		return entry, err
	}

	if resp.Item == nil {
		return entry, fmt.Errorf("Alias does not exist in index")
	}

	var a dynamoDBAlias
	err = dynamodbattribute.UnmarshalMap(resp.Item, &a)

	if err != nil {
		return entry, err
	}

	// Get on the Entries table
	return i.Get(a.Id)

}
func (i *DynamoDBIndex) UnAlias(alias string) error {
	// Deletes from alias table
	key := map[string]*dynamodb.AttributeValue{
		"Group": {
			S: aws.String(i.group),
		},
		"Alias": {
			S: aws.String(alias),
		},
	}

	params := (&dynamodb.DeleteItemInput{}).
		SetTableName(i.aliasesTable()).
		SetKey(key)

	_, err := i.ddb.DeleteItem(params)
	return err

}
func (i *DynamoDBIndex) Relate(a, b string) error {
	// Checks that both exist
	if i.Exists(a) != true || i.Exists(b) != true {
		return fmt.Errorf("Both entries must exist to be related")
	}

	// Puts to relations table
	r := dynamoDBRelation{
		A: i.group + "-" + a,
		B: b,
	}

	av, err := dynamodbattribute.MarshalMap(r)
	if err != nil {
		return err
	}

	params := (&dynamodb.PutItemInput{}).
		SetTableName(i.relationsTable()).
		SetItem(av)

	_, err = i.ddb.PutItem(params)
	return err
}

func (i *DynamoDBIndex) UnRelate(a, b string) error {
	// Deletes from Relations table
	r := dynamoDBRelation{
		A: i.group + "-" + a,
		B: b,
	}

	av, err := dynamodbattribute.MarshalMap(r)
	if err != nil {
		return err
	}

	params := (&dynamodb.DeleteItemInput{}).
		SetTableName(i.relationsTable()).
		SetKey(av)

	_, err = i.ddb.DeleteItem(params)
	return err

}
func (i *DynamoDBIndex) Relations(id string) []string {
	// Queries relations table

	values := map[string]*dynamodb.AttributeValue{
		":a": {
			S: aws.String(i.group + "-" + id),
		},
	}

	params := (&dynamodb.QueryInput{}).
		SetExpressionAttributeValues(values).
		SetKeyConditionExpression("A = :a").
		SetTableName(i.relationsTable())

	results := make([]string, 0)

	err := i.ddb.QueryPages(params,
		func(page *dynamodb.QueryOutput, lastPage bool) bool {
			var relations []dynamoDBRelation

			err := dynamodbattribute.UnmarshalListOfMaps(page.Items, &relations)
			if err != nil {
				log.Fatal("Error unmarshaling relations query: ", err)
			}

			for i := 0; i < len(relations); i++ {
				results = append(results, relations[i].B)
			}

			return !lastPage
		})

	if err != nil {
		log.Fatal("Error querying relations table: ", err)
	}

	return results
}
