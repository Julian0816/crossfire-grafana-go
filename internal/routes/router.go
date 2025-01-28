package routes

import (
	"github.com/gin-gonic/gin"
	"crossfire-grafana/internal/handlers"
)

// SetupRouter configures the Gin router.
func SetupRouter(projectID, databaseID string) *gin.Engine {
	router := gin.Default()

	// Base route
	router.GET("/", handlers.HomeHandler)

	// Restaurants cache route
	router.GET("/restaurants-cache", func(c *gin.Context) {
		handlers.RestaurantsCacheHandler(c, projectID, databaseID)
	})

	// Latest orders route
	router.GET("/latest-orders", func(c *gin.Context) {
		handlers.LatestOrdersHandler(c, projectID, databaseID)
	})

	// Dead letters route
	router.GET("/dead-letters-specific", func(c *gin.Context) {
		handlers.DeadLettersHandler(c, projectID, databaseID)
	})

	return router
}
