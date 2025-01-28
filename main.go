package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2/google"
	"github.com/joho/godotenv"
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
	url := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/%s/documents/%s", projectID, databaseID, collection)

	token, err := getFirestoreAccessToken()
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

func fetchDocumentsFromFirestoreWithSubcollection(projectID, databaseID, subCollection string) ([]FirestoreDocument, error) {
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

    token, err := getFirestoreAccessToken()
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

func fetchSpecificDocumentsFromFirestore(projectID, databaseID, parentCollection, subCollection string) ([]map[string]interface{}, error) {
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

	token, err := getFirestoreAccessToken()
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

	// Extract and return data
	var documents []map[string]interface{}
	for _, res := range result {
		if res.Document.Fields != nil {
			documents = append(documents, map[string]interface{}{
				"name":        res.Document.Name,
				"fields":      res.Document.Fields,
				"subCategory": subCollection, // Include subCollection for context
			})
		}
	}

	return documents, nil
}



// setupRouter configures the Gin router.
func setupRouter(projectID, databaseID string) *gin.Engine {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Server is running"})
	})

	// Fetch from the "restaurants" collection
	router.GET("/restaurants-cache", func(c *gin.Context) {
		restaurantsCollection := "restaurants"

		documents, err := fetchDocumentsFromFirestore(projectID, databaseID, restaurantsCollection)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"message":  "Documents fetched successfully from restaurants",
			"documents": documents,
		})
	})

// Fetch from the "latest-orders" collection
// Fetch from the "latest-orders" collection
router.GET("/latest-orders", func(c *gin.Context) {
    subCollectionID := c.Query("subCollection") // Get subcollection ID from query params (e.g., ?subCollection=I001)
    if subCollectionID == "" {
        c.JSON(400, gin.H{"error": "subCollection query parameter is required"})
        return
    }

    documents, err := fetchDocumentsFromFirestoreWithSubcollection(projectID, databaseID, subCollectionID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Process the documents to include combined fields
    var processedDocuments []map[string]interface{}
    for _, doc := range documents {
        fields := doc.Fields

        // Extract fields for combinedField
        var orderNumber, createdAt, datePosted string
        if orderNumberField, ok := fields["orderNumber"]; ok {
            orderNumber = orderNumberField.(map[string]interface{})["stringValue"].(string)
        }
        if createdAtField, ok := fields["createdAt"]; ok {
            createdAt = createdAtField.(map[string]interface{})["stringValue"].(string)
        }
        if datePostedField, ok := fields["datePosted"]; ok {
            datePosted = datePostedField.(map[string]interface{})["stringValue"].(string)
        }

        // Combine fields into a single string
        combinedField := fmt.Sprintf("%s - %s - %s - %s", subCollectionID, orderNumber, createdAt, datePosted)

        processedDocuments = append(processedDocuments, map[string]interface{}{
            "name":          doc.Name,
            "fields":        doc.Fields,
            "combinedField": combinedField,
        })
    }

    c.JSON(200, gin.H{
        "message":  "Documents fetched successfully",
        "documents": processedDocuments,
    })
})




	// Fetch from the "dead-letters" collection
	router.GET("/dead-letters-specific", func(c *gin.Context) {
    parentCollection := "dead-letters/NANALL"
    subCollection := c.Query("subCollection") // Example: `2024-12-16`

    if subCollection == "" {
        c.JSON(400, gin.H{"error": "subCollection query parameter is required"})
        return
    }

    documents, err := fetchSpecificDocumentsFromFirestore(projectID, databaseID, parentCollection, subCollection)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Process the documents to include combined fields
    var processedDocuments []map[string]interface{}
    for _, doc := range documents {
        fields := doc["fields"].(map[string]interface{})
        originalPayload := fields["originalPayload"].(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})
        storeOrders := originalPayload["StoreOrders"].(map[string]interface{})["arrayValue"].(map[string]interface{})["values"].([]interface{})

        for _, storeOrder := range storeOrders {
            orderFields := storeOrder.(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})
            combinedField := fmt.Sprintf("%s - %s - %s - %s - %s",
                originalPayload["OrderNumber"].(map[string]interface{})["stringValue"],
                orderFields["BillTo"].(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})["State"].(map[string]interface{})["stringValue"],
                orderFields["BillTo"].(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})["StoreCode"].(map[string]interface{})["stringValue"],
                orderFields["BillTo"].(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})["Suburb"].(map[string]interface{})["stringValue"],
                fields["errorMessage"].(map[string]interface{})["stringValue"],
            )

            processedDocuments = append(processedDocuments, map[string]interface{}{
                "combinedField": combinedField,
                "name":          doc["name"],
                "fields":        fields,
            })
        }
    }

    c.JSON(200, gin.H{
        "message":  "Documents fetched successfully",
        "documents": processedDocuments,
    })
})





	return router
}

func main() {

	err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

	projectID := os.Getenv("PROJECT_ID")
    databaseID := os.Getenv("DATABASE_ID")

	// Set up the HTTP server
	router := setupRouter(projectID, databaseID)

	log.Println("Server is running on port 4000")
	if err := router.Run(":4000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
