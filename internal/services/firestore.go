package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2/google"
)

// FirestoreDocument represents a Firestore document.
type FirestoreDocument struct {
	Name   string                 `json:"name"`
	Fields map[string]interface{} `json:"fields"`
}

// GetFirestoreAccessToken generates an OAuth token for Firestore.
func GetFirestoreAccessToken() (string, error) {
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/datastore")
	if err != nil {
		return "", fmt.Errorf("failed to find default credentials: %v", err)
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %v", err)
	}
	return token.AccessToken, nil
}


func FetchDocumentsFromFirestore(projectID, databaseID, collection string) ([]FirestoreDocument, error) {
	url := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/%s/documents/%s", projectID, databaseID, collection)

	var allDocuments []FirestoreDocument
	var nextPageToken string

	for {
		// Construct the URL with pagination if a next page token exists
		requestURL := url
		if nextPageToken != "" {
			requestURL = fmt.Sprintf("%s?pageToken=%s", url, nextPageToken)
		}

		// Get Firestore access token
		token, err := GetFirestoreAccessToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get access token: %v", err)
		}

		// Create the request
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		// Make the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("firestore API returned error: %s", resp.Status)
		}

		// Decode the response
		var result struct {
			Documents      []FirestoreDocument `json:"documents"`
			NextPageToken  string              `json:"nextPageToken"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v", err)
		}

		// Append the documents from this page
		allDocuments = append(allDocuments, result.Documents...)

		// Check if there is another page of documents
		if result.NextPageToken == "" {
			break
		}
		nextPageToken = result.NextPageToken
	}

	return allDocuments, nil
}


// FetchDocumentsFromFirestoreWithSubcollection queries a Firestore subcollection.
func FetchDocumentsFromFirestoreWithSubcollection(projectID, databaseID, subCollection string) ([]FirestoreDocument, error) {
	url := fmt.Sprintf(
		"https://firestore.googleapis.com/v1/projects/%s/databases/%s/documents:runQuery",
		projectID, databaseID,
	)

	payload := fmt.Sprintf(`{
        "structuredQuery": {
            "from": [{"collectionId": "%s", "allDescendants": true}]
        }
    }`, subCollection)

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	token, err := GetFirestoreAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Firestore API returned error: %s", resp.Status)
	}

	var result []struct {
		Document FirestoreDocument `json:"document"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	var documents []FirestoreDocument
	for _, res := range result {
		documents = append(documents, res.Document)
	}

	return documents, nil
}

// FetchSpecificDocumentsFromFirestore queries a specific Firestore collection.
func FetchSpecificDocumentsFromFirestore(projectID, databaseID, parentCollection, subCollection string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf(
		"https://firestore.googleapis.com/v1/projects/%s/databases/%s/documents:runQuery",
		projectID, databaseID,
	)

	payload := fmt.Sprintf(`{
		"structuredQuery": {
			"from": [{"collectionId": "%s", "allDescendants": true}]
		}
	}`, subCollection)

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	token, err := GetFirestoreAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Firestore API returned error: %s", resp.Status)
	}

	var result []struct {
		Document struct {
			Name   string                 `json:"name"`
			Fields map[string]interface{} `json:"fields"`
		} `json:"document"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	var documents []map[string]interface{}
	for _, res := range result {
		if res.Document.Fields != nil {
			documents = append(documents, map[string]interface{}{
				"name":        res.Document.Name,
				"fields":      res.Document.Fields,
				"subCategory": subCollection,
			})
		}
	}

	return documents, nil
}
