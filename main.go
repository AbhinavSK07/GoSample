package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// 1. The Struct Model
type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// 2. In-Memory Database & Mutex for thread safety
var (
	store  = make(map[int]Item)
	nextID = 1
	mu     sync.Mutex
)

func main() {
	// Initialize Gin router
	r := gin.Default()

	// 3. Define Routes using Gin
	r.POST("/items", createItem)
	r.GET("/items", getItems)
	r.GET("/items/:id", getItemByID)
	r.PUT("/items/:id", updateItem)
	r.DELETE("/items/:id", deleteItem)

	// Fetch the PORT from the environment (Render will set this)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback for local development
	}

	log.Printf("Server starting on port %s", port)

	// Start the Gin server
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// ==========================================
// HANDLER FUNCTIONS
// ==========================================

// CREATE
func createItem(c *gin.Context) {
	var item Item

	// c.ShouldBindJSON automatically parses the request body
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	mu.Lock()
	item.ID = nextID
	nextID++
	store[item.ID] = item
	mu.Unlock()

	// c.JSON automatically sets Content-Type to application/json
	c.JSON(http.StatusCreated, item)
}

// READ ALL
func getItems(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	items := make([]Item, 0, len(store))
	for _, item := range store {
		items = append(items, item)
	}

	c.JSON(http.StatusOK, items)
}

// READ ONE
func getItemByID(c *gin.Context) {
	// Gin uses c.Param to extract path variables defined with ":"
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	mu.Lock()
	item, exists := store[id]
	mu.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// UPDATE
func updateItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var updatedData Item
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	item, exists := store[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Update fields
	item.Name = updatedData.Name
	item.Price = updatedData.Price
	store[id] = item

	c.JSON(http.StatusOK, item)
}

// DELETE
func deleteItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := store[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	delete(store, id)

	// Added message after deletion!
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Item deleted successfully",
	})
}
