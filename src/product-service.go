package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)


type Product struct {
	ProductID    int    `json:"product_id" binding:"required,min=1"`
	Name 	     string `json:"name"`
	SKU          string `json:"sku" binding:"required,min=1,max=100"`
	Manufacturer string `json:"manufacturer" binding:"required,min=1,max=200"`
	CategoryID   int    `json:"category_id" binding:"required,min=1"`
	Weight       int    `json:"weight" binding:"min=0"`
	SomeOtherID  int    `json:"some_other_id" binding:"required,min=1"`
}

var products sync.Map

var brands = []string{"Alpha", "Beta", "Gamma", "Delta", "Nova"}

var categoryNames = []string{"Electronics","Books","Home","Sports","Toys","Clothing", "Garden","Beauty"}

func generateProducts(n int){
	for i := 1; i <= n; i++{
		brand := brands[i % len(brands)]
		products.Store(i, Product{
			ProductID:    i,
			Name:         fmt.Sprintf("Product %s %d", brand, i),
			SKU:          fmt.Sprintf("SKU-%d", i),
			Manufacturer: brand,
			CategoryID:   (i % len(categoryNames)) + 1,
			Weight:       100 + i%900,
			SomeOtherID:  i,
		})
	}
}
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

	v, found := products.Load(id)

	if !found {
		c.IndentedJSON(http.StatusNotFound, gin.H{
  			"error": "INVALID_INPUT",
  			"message": "The provided input data is invalid",
  			"details": "Product ID must be a positive integer",
		})
		return
	}
	
	product := v.(Product)

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


	if _, found := products.Load(id); !found {
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


	products.Store(id, productDetail)

    c.IndentedJSON(http.StatusOK, gin.H{
        "message": "Product details added sucessfully!",
    })
}