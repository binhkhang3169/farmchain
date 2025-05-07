package services

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go-backend/db"
	"go-backend/models"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Role string
}

type ClientMessage struct {
	Type    string  `json:"type"`
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Price   float64 `json:"price"`
}

type PriceResponse struct {
	SellerPrice float64 `json:"seller_price"`
	BuyerPrice  float64 `json:"buyer_price"`
	FairPrice   float64 `json:"fair_price"`
	Suggestion  string  `json:"suggestion"`
	Type        string  `json:"type"`
}

var (
	clients    = make(map[string]*websocket.Conn)
	prices     = make(map[string]float64)
	rolesReady = make(map[string]bool)
	allClients = make(map[*Client]bool)
	mutex      = &sync.Mutex{}
	Broadcast  = make(chan models.ChatMessage)
)

func HandleMessages() {
	for {
		msg := <-Broadcast
		saveToDatabase(msg)
		broadcastToClients(msg)
	}
}

func broadcastToClients(msg models.ChatMessage) {
	mutex.Lock()
	defer mutex.Unlock()
	for client := range allClients {
		if err := client.Conn.WriteJSON(msg); err != nil {
			log.Println("❌ Error sending message:", err)
			client.Conn.Close()
			delete(allClients, client)
		}
	}
}

func saveToDatabase(msg models.ChatMessage) {
	if err := db.DB.Create(&msg).Error; err != nil {
		log.Println("❌ Failed to save message to DB:", err)
	} else {
		log.Println("✅ Message saved to DB:", msg)
	}
}

func HandleWebSocket(conn *websocket.Conn) {
	defer conn.Close()
	client := &Client{Conn: conn}
	mutex.Lock()
	allClients[client] = true
	mutex.Unlock()

	var role string

	for {
		var data ClientMessage
		if err := conn.ReadJSON(&data); err != nil {
			log.Println("❌ Read error:", err)
			break
		}
		log.Printf("📥 Nhận dữ liệu từ client: %+v\n", data)

		switch data.Type {
		case "role":
			if data.Role == "seller" || data.Role == "buyer" {
				role = data.Role
				client.Role = role
				mutex.Lock()
				clients[role] = conn
				mutex.Unlock()
				log.Println("👤 Role set:", role)
			}
		case "message":
			if role == "" {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Role chưa được thiết lập"}`))
				continue
			}

			price, err := strconv.ParseFloat(data.Content, 64)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Giá không hợp lệ"}`))
				continue
			}

			mutex.Lock()
			prices[role] = price
			rolesReady[role] = true
			sellerPrice, sellerOK := prices["seller"]
			buyerPrice, buyerOK := prices["buyer"]
			sellerConn := clients["seller"]
			buyerConn := clients["buyer"]
			mutex.Unlock()

			log.Printf("💾 Lưu giá %s: %.2f", role, price)
			log.Printf("📊 Trạng thái rolesReady: seller=%v, buyer=%v", rolesReady["seller"], rolesReady["buyer"])
			log.Printf("📊 Trạng thái prices: seller=%.2f, buyer=%.2f", sellerPrice, buyerPrice)

			msg := models.ChatMessage{
				SenderRole: role,
				Content:    data.Content,
			}
			msg.Type = "chat"
			Broadcast <- msg

			if sellerOK && buyerOK && rolesReady["seller"] && rolesReady["buyer"] {
				log.Println("🚀 Đủ điều kiện gọi AI, gửi yêu cầu...")

				resp, err := callAIService(sellerPrice, buyerPrice)
				if err != nil {
					log.Println("❌ Lỗi khi gọi AI:", err)
					continue
				}
				resp.Type = "ai_response"

				if sellerConn != nil {
					sellerConn.WriteJSON(resp)
				}
				if buyerConn != nil {
					buyerConn.WriteJSON(resp)
				}

				mutex.Lock()
				delete(prices, "seller")
				delete(prices, "buyer")
				delete(rolesReady, "seller")
				delete(rolesReady, "buyer")
				mutex.Unlock()
			}
		}
	}

	mutex.Lock()
	delete(allClients, client)
	mutex.Unlock()
	log.Println("🔴 WebSocket client disconnected")
}

func callAIService(seller float64, buyer float64) (*PriceResponse, error) {
	body := map[string]float64{
		"seller_price": seller,
		"buyer_price":  buyer,
	}
	jsonData, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "http://flask-predictor:5000/predict", bytes.NewBuffer(jsonData)) // Sử dụng tên service thay vì localhost khi containerized
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	// Thiết lập timeout cho HTTP request
	client.Timeout = 10 * time.Second

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
