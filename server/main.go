package main

import (
	"go-backend/db"
	"go-backend/models"
	"go-backend/routes"
	"go-backend/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Kết nối DB & AutoMigrate các bảng
	db.Connect()
	db.DB.AutoMigrate(&models.AgriculturalArea{}, &models.ChatMessage{})

	// Chạy service xử lý message WebSocket
	go services.HandleMessages()

	// Khởi tạo router
	router := gin.Default()
	router.Use(corsMiddleware())

	// Đăng ký tất cả các route
	routes.RegisterRoutes(router)

	// Chạy server
	router.Run(":8080")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
