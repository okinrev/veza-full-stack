package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Import des protobuf générés
	chatpb "github.com/okinrev/veza-web-app/internal/grpc/generated/chat"
	streampb "github.com/okinrev/veza-web-app/internal/grpc/generated/stream"
)

func main() {
	// Logger simple
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("🚀 Veza Backend - Phase 2 : Test intégration gRPC")

	// Router Gin
	router := gin.Default()

	// Clients gRPC (optionnels)
	var chatClient chatpb.ChatServiceClient
	var streamClient streampb.StreamServiceClient

	// Tentative connexion Chat Server
	if conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials())); err == nil {
		chatClient = chatpb.NewChatServiceClient(conn)
		logger.Info("✅ Chat gRPC connecté")
	} else {
		logger.Warn("⚠️ Chat gRPC indisponible", zap.Error(err))
	}

	// Tentative connexion Stream Server
	if conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials())); err == nil {
		streamClient = streampb.NewStreamServiceClient(conn)
		logger.Info("✅ Stream gRPC connecté")
	} else {
		logger.Warn("⚠️ Stream gRPC indisponible", zap.Error(err))
	}

	// Route de santé
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"phase":  "gRPC Integration Test",
			"grpc": gin.H{
				"chat_connected":   chatClient != nil,
				"stream_connected": streamClient != nil,
			},
		})
	})

	// Test Chat gRPC
	router.POST("/test/chat", func(c *gin.Context) {
		if chatClient == nil {
			c.JSON(503, gin.H{"error": "Chat service unavailable"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := &chatpb.CreateRoomRequest{
			Name:       "Test Room",
			CreatedBy:  1,
			Type:       0,
			Visibility: 0,
		}

		resp, err := chatClient.CreateRoom(ctx, req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"room":    resp.Room,
		})
	})

	// Test Stream gRPC
	router.POST("/test/stream", func(c *gin.Context) {
		if streamClient == nil {
			c.JSON(503, gin.H{"error": "Stream service unavailable"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := &streampb.CreateStreamRequest{
			Title:      "Test Stream",
			StreamerId: 1,
			Category:   0,
		}

		resp, err := streamClient.CreateStream(ctx, req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"stream":  resp.Stream,
		})
	})

	// Démarrage serveur
	port := "8080"
	logger.Info("🌟 Serveur test gRPC démarré", zap.String("port", port))

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Erreur serveur:", err)
	}
}
