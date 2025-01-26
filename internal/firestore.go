package internal

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// CreateFirestoreClient initializes a Firestore client
func CreateFirestoreClient(ctx context.Context, projectID string) *firestore.Client {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	return client
}

// FetchCollectionData fetches all documents from a Firestore collection
func FetchCollectionData(client *firestore.Client, collectionName string) ([]map[string]interface{}, error) {
	ctx := context.Background()
	var results []map[string]interface{}

	fmt.Printf("Fetching documents from collection: %s\n", collectionName)

	iter := client.Collection(collectionName).Documents(ctx)
	defer iter.Stop()

	docCount := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error iterating through documents: %w", err)
		}

		docCount++
		fmt.Printf("Found document: %s => %v\n", doc.Ref.ID, doc.Data())
		results = append(results, map[string]interface{}{
			"id":   doc.Ref.ID,
			"data": doc.Data(),
		})
	}

	fmt.Printf("Total documents fetched: %d\n", docCount)
	return results, nil
}

