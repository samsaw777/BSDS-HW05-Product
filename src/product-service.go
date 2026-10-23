package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)


type Product struct {
	ProductID    int    `json:"product_id" binding:"required,min=1"`
	SKU          string `json:"sku" binding:"required,min=1,max=100"`
	Manufacturer string `json:"manufacturer" binding:"required,min=1,max=200"`
	CategoryID   int    `json:"category_id" binding:"required,min=1"`
	Weight       int    `json:"weight" binding:"min=0"`
	SomeOtherID  int    `json:"some_other_id" binding:"required,min=1"`
}

var (
	products = map[int]Product{
		1100: {ProductID: 1100, SKU: "GO-1100", Manufacturer: "Go Server", CategoryID: 1, Weight: 500, SomeOtherID: 10},
		1200: {ProductID: 1200, SKU: "TS-1200", Manufacturer: "Typescript", CategoryID: 20, Weight: 100, SomeOtherID: 14},
		1300: {ProductID: 1300, SKU: "AMZ-1300", Manufacturer: "Amazon", CategoryID: 13, Weight: 200, SomeOtherID: 20},
		1400: {ProductID: 1400, SKU: "NODE-1400", Manufacturer: "Node Server", CategoryID: 10, Weight: 200, SomeOtherID: 11},
	}
	mu sync.RWMutex
)

func getProducts(c *gin.Context){

	if c.GetHeader("X-Simulate-Error") == "true"{
		panic("Simulated internal error!")
	}

	id, err := strconv.Atoi(c.Param("productId"))

	if err != nil || id < 1{
		c.IndentedJSON(http.StatusNotFound, gin.H{
			"error": "INVALID_INPUT",
  			"message": "The provided input data is invalid",
  			"details": "Product ID must be a positive integer",
		})
		return
	}

	mu.RLock()
	product, found := products[id]
	mu.RUnlock()

	if !found {
		c.IndentedJSON(http.StatusNotFound, gin.H{
  			"error": "INVALID_INPUT",
  			"message": "The provided input data is invalid",
  			"details": "Product ID must be a positive integer",
		})
		return
	}

	c.IndentedJSON(http.StatusOK, product)
}

func updateProductsDetails(c *gin.Context) {
	if c.GetHeader("X-Simulate-Error") == "true" {
		panic("Simulated internal Error!")
	}

	id, err := strconv.Atoi(c.Param("productId"))
	if err != nil || id < 1 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_INPUT",
			"message": "The provided input data is invalid",
			"details": "Product ID must be a positive integer",
		})
		return
	}


	mu.RLock()
	_, found := products[id]
	mu.RUnlock()
	if !found {
		c.IndentedJSON(http.StatusNotFound, gin.H{
			"error":   "NOT_FOUND",
			"message": "Product not found",
			"details": "No product exists with ID " + strconv.Itoa(id),
		})
		return
	}


	var productDetail Product
	if err := c.ShouldBindJSON(&productDetail); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_INPUT",
			"message": "The provided input data is invalid",
			"details": err.Error(),
		})
		return
	}


	if productDetail.ProductID != id {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_INPUT",
			"message": "The provided input data is invalid",
			"details": "product_id in body must match productId in URL",
		})
		return
	}


	mu.Lock()
	products[id] = productDetail
	mu.Unlock()

  c.Status(http.StatusNoContent)

}