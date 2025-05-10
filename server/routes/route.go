package routes

import (
	"encoding/json"
	"go-backend/db"
	"go-backend/models"
	"go-backend/services"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// PriceData represents the price data structure
type PriceData struct {
	Day   string  `json:"day"`
	Price float64 `json:"price"`
}

// PriceResponse is the response structure from the Python API
type PriceResponse struct {
	History []PriceData `json:"history"`
	Predict []PriceData `json:"predictions"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // or check CORS based on environment
	},
}

func RegisterRoutes(router *gin.Engine) {
	router.POST("/areas", createArea)
	router.GET("/areas", getAreas)
	router.GET("/ws/chat", websocketHandler)
	router.GET("/api/price-history", getPriceHistory)
	router.GET("/api/chat-history", getChatHistory)
	
	// Start the message handling goroutine
	go services.HandleMessages()
}

func websocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("❌ Could not upgrade WebSocket:", err)
		return
	}
	
	log.Println("✅ New WebSocket connection established")
	
	// Handle the WebSocket connection
	go services.HandleWebSocket(conn)
}

func createArea(c *gin.Context) {
	var area models.AgriculturalArea
	if err := c.ShouldBindJSON(&area); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.DB.Create(&area).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create area"})
		return
	}
	c.JSON(http.StatusOK, area)
}

func getAreas(c *gin.Context) {
	var areas []models.AgriculturalArea
	if err := db.DB.Find(&areas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get areas"})
		return
	}
	c.JSON(http.StatusOK, areas)
}

func getChatHistory(c *gin.Context) {
	var messages []models.ChatMessage
	if err := db.DB.Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat history"})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func getPriceHistory(c *gin.Context) {
	// URL of Python API
	url := "http://flask-predictor:5000/price-history"

	// Call the Python API
	resp, err := http.Get(url)
	if err != nil {
		log.Println("❌ Error calling Python API:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price history"})
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("❌ Error reading response:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Python API returned error code: %d, body: %s", resp.StatusCode, string(body))
		c.JSON(resp.StatusCode, gin.H{"error": string(body)})
		return
	}

	// Parse JSON from Python API
	var priceData PriceResponse
	if err := json.Unmarshal(body, &priceData); err != nil {
		log.Println("❌ Error parsing JSON:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	// Return data to client
	c.JSON(http.StatusOK, priceData)
}