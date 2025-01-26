package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2/google"
)

// FirestoreDocument represents a Firestore document.
type FirestoreDocument struct {
	Name   string                 `json:"name"`
	Fields map[string]interface{} `json:"fields"`
}

// getFirestoreAccessToken generates an OAuth token for Firestore.
func getFirestoreAccessToken() (string, error) {
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

// fetchDocumentsFromFirestore queries the Firestore database using the REST API.
func fetchDocumentsFromFirestore(projectID, databaseID, collection string) ([]FirestoreDocument, error) {
	// Construct the URL for the Firestore REST API
	url := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/%s/documents/%s", projectID, databaseID, collection)

	// Get the OAuth2 token
	token, err := getFirestoreAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Send the HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Firestore API returned error: %s", resp.Status)
	}

	// Decode the response
	var result struct {
		Documents []FirestoreDocument `json:"documents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return result.Documents, nil
}

// setupRouter configures the Gin router.
func setupRouter(projectID, databaseID, collection string) *gin.Engine {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Server is running"})
	})

	router.GET("/restaurants-cache", func(c *gin.Context) {
		documents, err := fetchDocumentsFromFirestore(projectID, databaseID, collection)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"message":  "Documents fetched successfully",
			"documents": documents,
		})
	})

	router.GET("/latest-orders", func(c *gin.Context) {
		documents, err := fetchDocumentsFromFirestore(projectID, databaseID, collection)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"message":  "Documents fetched successfully",
			"documents": documents,
		})
	})

	return router
}

func main() {
	// Firestore configuration
	projectID := "prod-supply-chain-0b4688d1"
	databaseID := "crossfire-edi-db"
	collection := "restaurants"

	// Set up the HTTP server
	router := setupRouter(projectID, databaseID, collection)

	log.Println("Server is running on port 4000")
	if err := router.Run(":4000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
