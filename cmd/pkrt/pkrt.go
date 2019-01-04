package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/drocamor/packrat/index"
	"github.com/drocamor/packrat/store"
	"github.com/pd0mz/go-maidenhead"
	"github.com/rwcarlsen/goexif/exif"
)

var (
	prIndex    index.Index
	thumbStore store.Store
	origStore  store.Store
)

type putStoreAsyncResult struct {
	address store.Address
	err     error
}

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: pkrt [files]")
	}

	filenames := os.Args[1:]

	// Set up the index, the original store, and the thumbnail store
	// TODO make this configurable
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-west-2")}))
	prIndex = index.NewDynamoDBIndex(sess, "rocamora", "testPR")
	thumbStore = store.NewAWSStore(sess, "testPRThumbIndex", "testprthumbs")
	origStore = store.NewAWSStore(sess, "testPRStoreIndex", "testprstore")

	// Take a list of files from the args
	for i := 0; i < len(filenames); i++ {
		// Process each file in the list
		err := process(filenames[i])
		if err != nil {
			log.Println("error was: ", err)
		}
	}

}

func isImage(filename string) bool {
	cmd := exec.Command("/usr/bin/identify", filename)
	err := cmd.Run()
	if err != nil {
		log.Printf("cmd.Run err: %v", err)
		return false
	}
	return true
}

func putStoreAsync(st store.Store, filename string) chan putStoreAsyncResult {
	c := make(chan putStoreAsyncResult)

	go func() {
		var result putStoreAsyncResult
		f, err := os.Open(filename)
		if err != nil {
			result.err = err
			c <- result
			return
		}
		defer f.Close()
		a, err := st.Put(f)
		result.address = a
		result.err = err
		c <- result
	}()

	return c
}

func createThumb(filename string) (string, error) {
	tmp, err := ioutil.TempFile("", "pkrt-thumb*.jpg")
	if err != nil {
		return "", err
	}
	log.Printf("thumb file is %q", tmp.Name())
	cmd := exec.Command("/usr/bin/convert", "-resize", "480000@", filename, tmp.Name())
	err = cmd.Run()
	if err != nil {
		log.Printf("cmd.Run err: %v", err)
	}

	return tmp.Name(), err
}

func tsAndLocation(filename string) (time.Time, string) {
	f, err := os.Open(filename)
	if err != nil {
		log.Printf("Error opening file for exif: %v", err)
		return time.Now(), ""
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		log.Printf("Can't stat file: %v", err)
		return time.Now(), ""
	}

	x, err := exif.Decode(f)
	if err != nil {
		log.Printf("Error decoding exif: %v", err)
		return stat.ModTime(), ""
	}

	var ts time.Time
	ts, err = x.DateTime()
	if err != nil {
		ts = stat.ModTime()
	}

	lat, long, err := x.LatLong()
	if err != nil {
		return ts, ""
	}

	p := maidenhead.NewPoint(lat, long)
	gs, err := p.GridSquare()
	if err != nil {
		log.Printf("Error getting gridsquare: %v", err)
		return ts, ""
	}

	return ts, gs
}

func process(filename string) error {
	// Determine if it is an image - if not, skip it and log it
	if !isImage(filename) {
		return fmt.Errorf("%q is not an image.")
	}

	// Start uploading the image
	origUploadChan := putStoreAsync(origStore, filename)

	// Create a thumbnail
	thumbFilename, err := createThumb(filename)
	defer os.Remove(thumbFilename)
	if err != nil {
		return err
	}

	// start uploading the thumbnail
	thumbUploadChan := putStoreAsync(thumbStore, thumbFilename)

	// Use exif to determine date and location
	ts, gridsquare := tsAndLocation(filename)

	// Wait for the thumbnail and the image to be uploaded
	origResult := <-origUploadChan
	thumbResult := <-thumbUploadChan

	if origResult.err != nil {
		return fmt.Errorf("Error uploading original: %v", err)
	}

	if thumbResult.err != nil {
		return fmt.Errorf("Error uploading original: %v", err)
	}

	// Create an entry in the index
	entry := index.Entry{
		Name:       filename,
		Timestamp:  ts,
		Gridsquare: gridsquare,
		Type:       "image",
		Addresses: map[string]store.Address{
			"orig":  origResult.address,
			"thumb": thumbResult.address,
		},
	}

	return prIndex.Add(entry)
}
