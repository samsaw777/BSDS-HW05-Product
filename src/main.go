package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main(){

	generateProducts(1000000)

	router := gin.New()
	router.Use(loggingMiddleware()) 
	router.Use(metricsMiddleware()) 
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Internal server error",
			"details": "An unexpected error occurred",
		})
	}))

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", healthHandler)
	router.GET("/products/search", searchHandler)
	router.GET("/products/:productId", getProducts)
	router.POST("/products/:productId/details", updateProductsDetails)

	router.Run(":8080")
}