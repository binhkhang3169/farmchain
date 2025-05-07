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

// PriceData biểu diễn cấu trúc dữ liệu giá
type PriceData struct {
	Day   string  `json:"day"`
	Price float64 `json:"price"`
}

// PriceResponse là cấu trúc phản hồi từ Python API
type PriceResponse struct {
	History []PriceData `json:"history"`
	Predict []PriceData `json:"predictions"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // hoặc kiểm tra CORS tùy vào môi trường
	},
}

func RegisterRoutes(router *gin.Engine) {
	router.POST("/areas", createArea)
	router.GET("/areas", getAreas)
	router.GET("/ws/chat", websocketHandler)
	router.GET("/api/price-history", getPriceHistory)
}

func websocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("❌ Không thể nâng cấp WebSocket:", err)
		return
	}
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
func getPriceHistory(c *gin.Context) {
	// URL của Python API
	url := "http://flask-predictor:5000/price-history"

	// Gọi API Python
	resp, err := http.Get(url)
	if err != nil {
		log.Println("❌ Lỗi khi gọi Python API:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price history"})
		return
	}
	defer resp.Body.Close()

	// Đọc phản hồi
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("❌ Lỗi khi đọc phản hồi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Kiểm tra mã trạng thái
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Python API trả về mã lỗi: %d, body: %s", resp.StatusCode, string(body))
		c.JSON(resp.StatusCode, gin.H{"error": string(body)})
		return
	}

	// Parse JSON từ Python API
	var priceData PriceResponse
	if err := json.Unmarshal(body, &priceData); err != nil {
		log.Println("❌ Lỗi khi parse JSON:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	// Trả về dữ liệu cho client
	c.JSON(http.StatusOK, priceData)
}
