// 📁 services/websocket_handler.go
package controllers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type ClientMessage struct {
	Type    string  `json:"type"`    // "message" hoặc "role"
	Role    string  `json:"role"`    // "seller" hoặc "buyer"
	Content string  `json:"content"` // chỉ gửi nếu type == message
	Price   float64 `json:"price"`   // giá nếu gửi giá
}

type PriceResponse struct {
	SellerPrice float64 `json:"seller_price"`
	BuyerPrice  float64 `json:"buyer_price"`
	FairPrice   float64 `json:"fair_price"`
	Suggestion  string  `json:"suggestion"`
	Type        string  `json:"type"` // "ai_response"
}

var (
	mu      sync.Mutex
	prices  = make(map[string]float64)
	clients = make(map[string]*websocket.Conn)
)

func HandleWebSocket(conn *websocket.Conn) {
	defer conn.Close()
	var role string

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("🛑 Mất kết nối WebSocket:", err)
			break
		}

		var data ClientMessage
		if err := json.Unmarshal(msg, &data); err != nil {
			log.Println("❌ Lỗi JSON:", err)
			continue
		}

		if data.Type == "role" && (data.Role == "seller" || data.Role == "buyer") {
			role = data.Role
			mu.Lock()
			clients[role] = conn
			mu.Unlock()
			continue
		}

		if data.Type == "message" && role != "" {
			price, err := parsePrice(data.Content)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Giá không hợp lệ"}`))
				continue
			}

			mu.Lock()
			prices[role] = price
			clients[role] = conn
			sellerPrice, hasSeller := prices["seller"]
			buyerPrice, hasBuyer := prices["buyer"]
			sellerConn := clients["seller"]
			buyerConn := clients["buyer"]
			mu.Unlock()

			// broadcast chat message
			chatMsg := map[string]interface{}{
				"type":        "chat",
				"sender_role": role,
				"content":     data.Content,
			}
			chatJSON, _ := json.Marshal(chatMsg)
			if sellerConn != nil {
				sellerConn.WriteMessage(websocket.TextMessage, chatJSON)
			}
			if buyerConn != nil {
				buyerConn.WriteMessage(websocket.TextMessage, chatJSON)
			}

			if hasSeller && hasBuyer && sellerConn != nil && buyerConn != nil {
				resp, err := callAIService(sellerPrice, buyerPrice)
				if err != nil {
					log.Println("❌ Lỗi khi gọi AI:", err)
					continue
				}
				resp.Type = "ai_response"
				responseJSON, _ := json.Marshal(resp)
				sellerConn.WriteMessage(websocket.TextMessage, responseJSON)
				buyerConn.WriteMessage(websocket.TextMessage, responseJSON)

				// reset state
				mu.Lock()
				delete(prices, "seller")
				delete(prices, "buyer")
				mu.Unlock()
			}
		}
	}
}

func parsePrice(content string) (float64, error) {
	return strconv.ParseFloat(content, 64)
}

func callAIService(seller float64, buyer float64) (*PriceResponse, error) {
	body := map[string]float64{
		"seller_price": seller,
		"buyer_price":  buyer,
	}
	jsonData, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "http://localhost:5000/predict", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

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
