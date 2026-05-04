package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
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
	mux := http.NewServeMux()

	// 3. Define Routes (Requires Go 1.22+)
	mux.HandleFunc("POST /items", createItem)
	mux.HandleFunc("GET /items", getItems)
	mux.HandleFunc("GET /items/{id}", getItemByID)
	mux.HandleFunc("PUT /items/{id}", updateItem)
	mux.HandleFunc("DELETE /items/{id}", deleteItem)

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// ==========================================
// HANDLER FUNCTIONS
// ==========================================

// CREATE
func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mu.Lock()
	item.ID = nextID
	nextID++
	store[item.ID] = item
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// READ ALL
func getItems(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	items := make([]Item, 0, len(store))
	for _, item := range store {
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// READ ONE
func getItemByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	item, exists := store[id]
	mu.Unlock()

	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// UPDATE
func updateItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedData Item
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	item, exists := store[id]
	if !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Update fields
	item.Name = updatedData.Name
	item.Price = updatedData.Price
	store[id] = item

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// DELETE
func deleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := store[id]; !exists {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	delete(store, id)

	w.WriteHeader(http.StatusNoContent)
}