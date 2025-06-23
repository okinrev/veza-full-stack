package chat

import (
	"time"

	"github.com/okinrev/veza-web-app/internal/database"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// GetRooms récupère la liste des salons
func (s *Service) GetRooms(userID int) ([]map[string]interface{}, error) {
	// TODO: Implémenter la récupération des salons depuis la BDD
	// Pour l'instant, retourner des données fictives
	rooms := []map[string]interface{}{
		{
			"id":          1,
			"name":        "general",
			"description": "Salon général",
			"is_private":  false,
			"created_at":  time.Now().Format(time.RFC3339),
		},
		{
			"id":          2,
			"name":        "random",
			"description": "Salon aléatoire",
			"is_private":  false,
			"created_at":  time.Now().Format(time.RFC3339),
		},
	}
	return rooms, nil
}

// GetRoomMessages récupère les messages d'un salon
func (s *Service) GetRoomMessages(roomID string, userID int) ([]map[string]interface{}, error) {
	// TODO: Implémenter la récupération des messages depuis la BDD
	// Pour l'instant, retourner des données fictives
	messages := []map[string]interface{}{
		{
			"id":        1,
			"room_id":   roomID,
			"user_id":   userID,
			"content":   "Hello room!",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
	return messages, nil
}

// SendRoomMessage envoie un message dans un salon
func (s *Service) SendRoomMessage(roomID string, userID int, content string) (map[string]interface{}, error) {
	// TODO: Implémenter l'envoi de message dans la BDD
	// Pour l'instant, retourner des données fictives
	message := map[string]interface{}{
		"id":        1,
		"room_id":   roomID,
		"user_id":   userID,
		"content":   content,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return message, nil
}

// GetDMMessages récupère les messages directs avec un utilisateur
func (s *Service) GetDMMessages(userID, targetUserID int) ([]map[string]interface{}, error) {
	// TODO: Implémenter la récupération des messages depuis la BDD
	// Pour l'instant, retourner des données fictives
	messages := []map[string]interface{}{
		{
			"id":        1,
			"from_user": userID,
			"to_user":   targetUserID,
			"content":   "Hello!",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
	return messages, nil
}

// SendDMMessage envoie un message direct à un utilisateur
func (s *Service) SendDMMessage(userID, targetUserID int, content string) (map[string]interface{}, error) {
	// TODO: Implémenter l'envoi de message dans la BDD
	// Pour l'instant, retourner des données fictives
	message := map[string]interface{}{
		"id":        1,
		"from_user": userID,
		"to_user":   targetUserID,
		"content":   content,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return message, nil
}
