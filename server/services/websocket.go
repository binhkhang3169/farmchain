package services

import (
	"regexp"
	"bytes"
    "encoding/json"
    "net/http"
	"log"
	"strconv"
	"sync"
	"go-backend/db"
	"go-backend/models"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Role string
	ID   string // Unique identifier for each client
}

type ClientMessage struct {
	Type       string `json:"type"`
	Role       string `json:"role"`
	Content    string `json:"content"`
	SenderRole string `json:"senderRole"`
}

type AIResponse struct {
	Type       string `json:"type"`
	SenderRole string `json:"senderRole"`
	FairPrice  string `json:"fair_price"`
	Suggestion string `json:"suggestion"`
	Content    string `json:"content,omitempty"`
}

var (
	clients    = make(map[string]*Client)     // Map of role -> client
	prices     = make(map[string]float64)      // Store prices by role
	allClients = make(map[*Client]bool)        // All connected clients
	mutex      = &sync.Mutex{}
	Broadcast  = make(chan models.ChatMessage) // Channel for broadcasting messages
	clientIDs  = 0                            // Counter for generating client IDs
)

// HandleMessages processes messages from the broadcast channel
func HandleMessages() {
	for {
		msg := <-Broadcast
		saveToDatabase(msg)
		broadcastToClients(msg)
		
	}
}

// broadcastToClients sends a message to all connected clients
func broadcastToClients(msg models.ChatMessage) {
	mutex.Lock()
	defer mutex.Unlock()
	
	log.Printf("Broadcasting message to %d clients: %+v", len(allClients), msg)
	
	for client := range allClients {
		if err := client.Conn.WriteJSON(msg); err != nil {
			log.Println("Error sending message:", err)
			client.Conn.Close()
			delete(allClients, client)
			// If this client was registered with a role, clean up that reference too
			if client.Role != "" && clients[client.Role] != nil && clients[client.Role].ID == client.ID {
				delete(clients, client.Role)
				delete(prices, client.Role)
			}
		}
	}
}

// saveToDatabase persists messages to the database
func saveToDatabase(msg models.ChatMessage) {
	if err := db.DB.Create(&msg).Error; err != nil {
		log.Println("Failed to save message to DB:", err)
	} else {
		log.Println("Message saved to DB:", msg)
	}
}

// HandleWebSocket manages a WebSocket connection
func HandleWebSocket(conn *websocket.Conn) {
	// Create a new client with a unique ID
	mutex.Lock()
	clientIDs++
	client := &Client{
		Conn: conn, 
		ID: strconv.Itoa(clientIDs),
	}
	allClients[client] = true
	mutex.Unlock()
	
	log.Printf("New client connected with ID: %s", client.ID)
	
	// Clean up on disconnect
	defer func() {
		log.Printf("Client disconnected: %s (role: %s)", client.ID, client.Role)
		mutex.Lock()
		delete(allClients, client)
		if client.Role != "" && clients[client.Role] != nil && clients[client.Role].ID == client.ID {
			delete(clients, client.Role)
			delete(prices, client.Role)
		}
		mutex.Unlock()
		conn.Close()
	}()
	
	// Message handling loop
	for {
		var data ClientMessage
		if err := conn.ReadJSON(&data); err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("Received data from client %s: %+v", client.ID, data)
		
		switch data.Type {
		case "role":
			handleRoleMessage(client, data)
		case "message":
			handleChatMessage(client, data)
		case "purchase":
			handlePurchaseMessage(client, data)
		case "price_update":
			handlePriceMessage(client, data)
		default:
			conn.WriteJSON(map[string]string{
				"type":  "error",
				"error": "Unknown message type",
			})
		}
	}
}

// handleRoleMessage processes role registration messages
func handleRoleMessage(client *Client, data ClientMessage) {
	if data.Role != "seller" && data.Role != "buyer" {
		client.Conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Invalid role. Must be 'buyer' or 'seller'",
		})
		return
	}
	
	mutex.Lock()
	// If this role is already registered by another client, replace it
	if oldClient, exists := clients[data.Role]; exists {
		if oldClient.ID != client.ID {
			log.Printf("Role '%s' was already registered by client %s, replacing with client %s", 
				data.Role, oldClient.ID, client.ID)
			// Don't delete from allClients, as the connection may still be alive
		}
	}
	
	// Update the client's role and register it
	client.Role = data.Role
	clients[data.Role] = client
	mutex.Unlock()
	
	// Confirm role registration
	response := map[string]string{
		"type":    "status",
		"status":  "success",
		"message": "Role set successfully",
	}
	
	if err := client.Conn.WriteJSON(response); err != nil {
		log.Printf("Error sending role confirmation: %v", err)
	} else {
		log.Printf("Client %s registered as %s", client.ID, data.Role)
	}
}

func extractPrice(content string) (int, bool) {
	re := regexp.MustCompile(`\d+`) // Tìm tất cả chuỗi số liên tiếp
	match := re.FindString(content) // Tìm chuỗi số đầu tiên
	if match != "" {
		price, err := strconv.Atoi(match)
		if err == nil {
			return price, true
		}
	}
	return 0, false
}


// handleChatMessage processes regular chat messages
func handleChatMessage(client *Client, data ClientMessage) {
	if client.Role == "" {
		client.Conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Please set your role before sending messages",
		})
		return
	}
	price, found := extractPrice(data.Content)
	

	// Prepare the message for broadcast
	msg := models.ChatMessage{
		Type:       "chat",
		SenderRole: client.Role,
		Content:    data.Content,
	}
	
	// Send to broadcast channel for distribution
	Broadcast <- msg

	if found {
		// Store the price for this role
		mutex.Lock()
		prices[client.Role] = float64(price)
		mutex.Unlock()
		checkAndTriggerAIResponse()	
	}
}

// handleChatMessage processes regular chat messages
func handlePurchaseMessage(client *Client, data ClientMessage) {
	if client.Role == "" {
		client.Conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Please set your role before sending messages",
		})
		return
	}	

	// Prepare the message for broadcast
	msg := models.ChatMessage{
		Type:       "purchase",
		SenderRole: client.Role,
		Content:    data.Content,
	}
	
	// Send to broadcast channel for distribution
	Broadcast <- msg
}


// handlePriceMessage processes price proposal/bid messages
func handlePriceMessage(client *Client, data ClientMessage) {
	if client.Role == "" {
		client.Conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Please set your role before sending price proposals",
		})
		return
	}
	
	// Parse the price
	price, err := strconv.ParseFloat(data.Content, 64)
	if err != nil {
		client.Conn.WriteJSON(map[string]string{
			"type":  "error",
			"error": "Invalid price format. Please send a valid number",
		})
		return
	}
	
	// Store the price for this role
	mutex.Lock()
	prices[client.Role] = price
	mutex.Unlock()
	
	// Create and broadcast the price message
	msg := models.ChatMessage{
		Type:       "price_sell",
		SenderRole: client.Role,
		Content:    data.Content,
	}
	
	// Send to broadcast channel
	Broadcast <- msg
}

// checkAndTriggerAIResponse checks if both buyer and seller have submitted prices
// and if so, triggers an AI response with negotiation advice using the Flask API
func checkAndTriggerAIResponse() {
    mutex.Lock()
    defer mutex.Unlock()
   
    // Check if we have prices from both buyer and seller
    buyerPrice, hasBuyerPrice := prices["buyer"]
    sellerPrice, hasSellerPrice := prices["seller"]
   
    if !hasBuyerPrice || !hasSellerPrice {
        // We don't have prices from both parties yet
        return
    }
   
    // Call the Flask API to get AI prediction
    apiURL := "http://flask-predictor:5000/predict"
    requestData := map[string]float64{
        "seller_price": sellerPrice,
        "buyer_price": buyerPrice,
    }
    
    jsonData, err := json.Marshal(requestData)
    if err != nil {
        log.Println("Error marshalling request data:", err)
        return
    }
    
    resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Println("Error calling AI API:", err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        log.Printf("API returned non-OK status: %d", resp.StatusCode)
        return
    }
    
    var apiResponse struct {
        SellerPrice float64 `json:"seller_price"`
        BuyerPrice  float64 `json:"buyer_price"`
        FairPrice   float64 `json:"fair_price"`
        Suggestion  string  `json:"suggestion"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
        log.Println("Error decoding API response:", err)
        return
    }
   
    // Format the fair price with thousands separator for display
    fairPriceStr := strconv.FormatFloat(apiResponse.FairPrice, 'f', 0, 64)
   
    // Create the AI response
    aiResponse := AIResponse{
        Type:       "ai_response",
        SenderRole: "ai",
        FairPrice:  fairPriceStr,
        Suggestion: apiResponse.Suggestion,
    }
   
    // Broadcast to all clients
    for client := range allClients {
        if err := client.Conn.WriteJSON(aiResponse); err != nil {
            log.Println("Error sending AI response:", err)
        }
    }
   
    // Optionally save AI message to database
    msg := models.ChatMessage{
        Type:       "ai_response",
        SenderRole: "ai",
        Content:    "Giá hợp lý: " + fairPriceStr + " đ/kg\n" + apiResponse.Suggestion,
    }
    saveToDatabase(msg)
   
    // Reset prices to allow for new negotiation round
    delete(prices, "buyer")
    delete(prices, "seller")
}