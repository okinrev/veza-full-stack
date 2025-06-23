package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/okinrev/veza-web-app/internal/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // En production, vérifier l'origine
	},
}

type Client struct {
	conn     *websocket.Conn
	userID   int
	username string
	send     chan []byte
}

type ChatManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
	jwtSecret  string
}

func NewChatManager(jwtSecret string) *ChatManager {
	return &ChatManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		jwtSecret:  jwtSecret,
	}
}

func (cm *ChatManager) Run() {
	for {
		select {
		case client := <-cm.register:
			cm.mutex.Lock()
			cm.clients[client] = true
			cm.mutex.Unlock()
			log.Printf("Client connecté: %s (ID: %d)", client.username, client.userID)

		case client := <-cm.unregister:
			cm.mutex.Lock()
			if _, ok := cm.clients[client]; ok {
				delete(cm.clients, client)
				close(client.send)
			}
			cm.mutex.Unlock()
			log.Printf("Client déconnecté: %s (ID: %d)", client.username, client.userID)

		case message := <-cm.broadcast:
			cm.mutex.Lock()
			for client := range cm.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(cm.clients, client)
				}
			}
			cm.mutex.Unlock()
		}
	}
}

func (cm *ChatManager) HandleWebSocket(c *gin.Context) {
	// Vérifier le token
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token manquant"})
		return
	}

	// Valider le token JWT
	claims, err := utils.ValidateJWT(token, cm.jwtSecret)
	if err != nil {
		log.Printf("Token invalide: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Extraire les informations de l'utilisateur
	userID := claims.UserID
	username := claims.Username

	// Mettre à niveau la connexion HTTP en WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Erreur lors de la mise à niveau WebSocket: %v", err)
		return
	}

	client := &Client{
		conn:     conn,
		userID:   userID,
		username: username,
		send:     make(chan []byte, 256),
	}

	cm.register <- client

	// Gérer la connexion dans une goroutine
	go func() {
		defer func() {
			cm.unregister <- client
			client.conn.Close()
		}()

		for {
			_, message, err := client.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Erreur de lecture WebSocket: %v", err)
				}
				break
			}

			// Traiter le message
			cm.broadcast <- message
		}
	}()

	// Gérer l'envoi des messages
	go func() {
		defer func() {
			client.conn.Close()
		}()

		for {
			select {
			case message, ok := <-client.send:
				if !ok {
					client.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := client.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(message)

				if err := w.Close(); err != nil {
					return
				}
			}
		}
	}()
}
