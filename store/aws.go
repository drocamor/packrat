package store

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	blobPrefix = "blobs/"
)

// AWSStore is a store that puts big objects into S3 and little ones into a DynamoDB table
type AWSStore struct {
	ddbSvc     *dynamodb.DynamoDB
	s3Svc      *s3.S3
	uploader   *s3manager.Uploader
	indexTable string // DynamoDB table to store index of objects
	bucket     string // S3 bucket for storing big objects
	userId     string
}

func NewAWSStore(sess *session.Session, indexTable, bucket string) *AWSStore {
	return &AWSStore{
		ddbSvc:     dynamodb.New(sess),
		s3Svc:      s3.New(sess),
		uploader:   s3manager.NewUploader(sess),
		indexTable: indexTable,
		bucket:     bucket,
	}
}

func (s *AWSStore) Put(r io.Reader) (Address, error) {
	var a Address
	// Create a temp file
	tmp, err := ioutil.TempFile("", "pkrt-staging")
	if err != nil {
		return a, err
	}
	defer os.Remove(tmp.Name())

	// Tee bytes from the reader to the temp file
	tee := io.TeeReader(r, tmp)

	// copy the tee to the hashing func
	h := sha256.New()
	length, err := io.Copy(h, tee)
	a.Size = length
	if err != nil {
		return a, err
	}

	score := fmt.Sprintf("%x", h.Sum(nil))
	a.Score = score

	// Use Describe to determine if the object is already in the index.
	describedA, err := s.Describe(score)
	log.Printf("describe err was: %v", err)
	// If it is, return the Address
	if err == nil {
		return describedA, err
	}

	// Rewind the tmpfile back to the beginning
	loc, err := tmp.Seek(0, io.SeekStart)
	if err != nil {
		return a, err
	}

	log.Printf("loc is : %d", loc)
	// Use the size to determine where to upload the data

	log.Printf("score is %q, length is %d", score, length)
	key := blobPrefix + score
	a.Location = fmt.Sprintf("s3://%s/%s", s.bucket, key)

	_, err = s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   tmp,
	})

	if err != nil {
		return a, err
	}

	err = s.putStoreIndex(a)

	// Return the Address
	return a, err

}

func (s *AWSStore) putStoreIndex(a Address) error {
	av, err := dynamodbattribute.MarshalMap(a)
	if err != nil {
		return err
	}

	params := (&dynamodb.PutItemInput{}).
		SetTableName(s.indexTable).
		SetConditionExpression("attribute_not_exists(Score)").
		SetItem(av)

	_, err = s.ddbSvc.PutItem(params)
	return err
}

func (s *AWSStore) Get(score string, w io.Writer) error {
	// Use Describe to determine the address of the blob
	addr, err := s.Describe(score)
	if err != nil {
		return err
	}

	// Use GetAddress to write the bytes to w
	return s.GetAddress(addr, w)

}

func resolveS3Uri(uri string) (bucket, key string, err error) {
	if strings.HasPrefix(uri, "s3://") != true {
		err = fmt.Errorf("Invalid s3 uri: %q", uri)
		return
	}

	shortUri := uri[5:]

	bucketAndKey := strings.SplitN(shortUri, "/", 2)

	if len(bucketAndKey) != 2 {
		err = fmt.Errorf("Invalid s3 uri: %q", uri)
		return
	}

	bucket = bucketAndKey[0]
	key = bucketAndKey[1]

	return
}

func (s *AWSStore) GetAddress(a Address, w io.Writer) error {
	bucket, key, err := resolveS3Uri(a.Location)
	if err != nil {
		return err
	}

	params := (&s3.GetObjectInput{}).
		SetBucket(bucket).
		SetKey(key)

	resp, err := s.s3Svc.GetObject(params)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(w, resp.Body)

	return err
}

func (s *AWSStore) Describe(score string) (Address, error) {
	a := Address{Score: score}
	// Look up the Address from the index table and return it.
	key := map[string]*dynamodb.AttributeValue{
		"Score": {
			S: aws.String(score),
		},
	}

	params := (&dynamodb.GetItemInput{}).
		SetTableName(s.indexTable).
		SetKey(key)

	resp, err := s.ddbSvc.GetItem(params)
	if err != nil {
		return a, err
	}

	if resp.Item == nil {
		return a, fmt.Errorf("Blob does not exist in store")
	}

	err = dynamodbattribute.UnmarshalMap(resp.Item, &a)

	return a, err
}
