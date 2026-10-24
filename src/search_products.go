package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxChecked = 100
	maxResults = 20
	maxQueryLen = 100
)

type SearchResponse struct {
	Products   []Product `json:"products"`
	TotalCount int       `json:"total_count"`
	SearchTime string    `json:"search_time"`
}

func categoryName(id int) string {
	if id < 1 || id > len(categoryNames){
		return ""
	}
	return categoryNames[id - 1]
}

func searchProducts(query string) ([]Product, int){
	q := strings.ToLower(query)
	results := make([]Product, 0, maxResults)
	totalCount := 0
	checked := 0

	products.Range(func (key, value any) bool  {
		p := value.(Product)

		nameMatch := strings.Contains(strings.ToLower(p.Name), q)
		categoryMatch := strings.Contains(strings.ToLower(categoryName(p.CategoryID)),q)

		if nameMatch || categoryMatch {
			totalCount++
			if len(results) < maxResults{
				results = append(results, p)
			}
		}
		
		checked++
		return checked < maxChecked
	})

	return results, totalCount
}

func searchHandler(c *gin.Context){
	q,ok := c.GetQuery("q")
	
	if !ok || len(q) < 1 || len(q) > maxQueryLen {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_INPUT",
			"message": "The provided input data is invalid",
			"details": "Query parameter q is required and must be 1-100 characters",
		})
		return
	}

	start := time.Now()
	results, total := searchProducts(q)
	elasped := time.Since(start)

		c.JSON(http.StatusOK, SearchResponse{
		Products:   results,
		TotalCount: total,
		SearchTime: fmt.Sprintf("%.3fms", float64(elasped.Microseconds())/1000.0),
	})
}

func healthHandler(c *gin.Context){
	c.JSON(http.StatusOK, gin.H{"status":"ok"})
}