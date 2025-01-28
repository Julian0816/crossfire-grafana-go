package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"crossfire-grafana/internal/services" 
)

// HomeHandler handles the base route.
func HomeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Server is running"})
}

// RestaurantsCacheHandler fetches data from the "restaurants" collection.
func RestaurantsCacheHandler(c *gin.Context, projectID, databaseID string) {
	restaurantsCollection := "restaurants"

	documents, err := services.FetchDocumentsFromFirestore(projectID, databaseID, restaurantsCollection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Documents fetched successfully from restaurants",
		"documents": documents,
	})
}

// LatestOrdersHandler fetches data from the "latest-orders" collection.
func LatestOrdersHandler(c *gin.Context, projectID, databaseID string) {
	subCollectionID := c.Query("subCollection")
	if subCollectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subCollection query parameter is required"})
		return
	}

	documents, err := services.FetchDocumentsFromFirestoreWithSubcollection(projectID, databaseID, subCollectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var processedDocuments []map[string]interface{}
	for _, doc := range documents {
		fields := doc.Fields
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

		combinedField := subCollectionID + " - " + orderNumber + " - " + createdAt + " - " + datePosted
		processedDocuments = append(processedDocuments, map[string]interface{}{
			"name":          doc.Name,
			"fields":        doc.Fields,
			"combinedField": combinedField,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Documents fetched successfully",
		"documents": processedDocuments,
	})
}

// DeadLettersHandler fetches data from the "dead-letters" collection.
func DeadLettersHandler(c *gin.Context, projectID, databaseID string) {
	parentCollection := "dead-letters/NANALL"
	subCollection := c.Query("subCollection")
	if subCollection == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subCollection query parameter is required"})
		return
	}

	documents, err := services.FetchSpecificDocumentsFromFirestore(projectID, databaseID, parentCollection, subCollection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var processedDocuments []map[string]interface{}
	for _, doc := range documents {
		fields := doc["fields"].(map[string]interface{})
		originalPayload := fields["originalPayload"].(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})
		storeOrders := originalPayload["StoreOrders"].(map[string]interface{})["arrayValue"].(map[string]interface{})["values"].([]interface{})

		for _, storeOrder := range storeOrders {
			orderFields := storeOrder.(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})
			combinedField := originalPayload["OrderNumber"].(map[string]interface{})["stringValue"].(string) + " - " +
				orderFields["BillTo"].(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})["State"].(map[string]interface{})["stringValue"].(string) + " - " +
				orderFields["BillTo"].(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})["StoreCode"].(map[string]interface{})["stringValue"].(string) + " - " +
				orderFields["BillTo"].(map[string]interface{})["mapValue"].(map[string]interface{})["fields"].(map[string]interface{})["Suburb"].(map[string]interface{})["stringValue"].(string) + " - " +
				fields["errorMessage"].(map[string]interface{})["stringValue"].(string)

			processedDocuments = append(processedDocuments, map[string]interface{}{
				"combinedField": combinedField,
				"name":          doc["name"],
				"fields":        fields,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Documents fetched successfully",
		"documents": processedDocuments,
	})
}
