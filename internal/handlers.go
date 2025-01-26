package internal

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

func GetMetricsHandler(c *gin.Context) {
    // Example data for Grafana
    data := []map[string]interface{}{
        {
            "timestamp": time.Now().Format(time.RFC3339),
            "metric":    "orders",
            "value":     100,
        },
        {
            "timestamp": time.Now().Add(10 * time.Minute).Format(time.RFC3339),
            "metric":    "orders",
            "value":     150,
        },
    }

    // Respond with JSON
    c.JSON(http.StatusOK, data)
}
