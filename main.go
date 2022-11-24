package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

func uploadLocalFileToGCSBucket(
	bucketName string,
	objectName string,
	fileName string,
) error {
	// create a context with cancel to ensure that all operations we do
	// using this context will be canceled when the function returns
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create GCS client. this will fetch credentials from the environment,
	// usually an environment variable called GOOGLE_APPLICATION_CREDENTIALS
	// containing the filename of a JSON OAuth private key of a service account
	// with permission to upload objects in the specified bucket. I'm not
	// sure how this should work for a mocked GCS API server, but this mock
	// server should reply something that will make the internal state of the
	// storage client be ready to make API calls
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("error creating GCS client: %w", err)
	}
	defer client.Close()

	// here we create a writer with the storage client to pass
	// to io.Copy() below in order to copy a local file into
	// a GCS object
	writer := client.
		Bucket(bucketName).
		Object(objectName).
		If(storage.Conditions{DoesNotExist: true}).
		NewWriter(ctx)
	defer writer.Close()

	// open a local file
	f, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error opening local file for reading: %w", err)
	}
	defer f.Close()

	// copy file into GCS object writer. if an error happens here,
	// it should never contain something like the following message:
	//
	// "googleapi: got HTTP response code 503 with body: Service Unavailable"
	//
	// but we observed that it does.
	if _, err := io.Copy(writer, f); err != nil {
		return fmt.Errorf("error uploading local file to GCS bucket: %w", err)
	}

	return nil
}

func main() {
	const bucketName = "support-case-snippet"
	const objectName = "license.txt"
	const fileName = "LICENSE"
	if err := uploadLocalFileToGCSBucket(bucketName, objectName, fileName); err != nil {
		log.Fatal(err)
	}
}
