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

// FetchDocumentsFromFirestore queries the Firestore database using the REST API.
func FetchDocumentsFromFirestore(projectID, databaseID, collection string) ([]FirestoreDocument, error) {
	url := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/%s/documents/%s", projectID, databaseID, collection)

	token, err := GetFirestoreAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("firestore API returned error: %s", resp.Status)
	}

	var result struct {
		Documents []FirestoreDocument `json:"documents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return result.Documents, nil
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
