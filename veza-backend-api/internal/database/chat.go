package database

import (
	"time"
)

type Message struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	RoomID    string    `json:"room_id,omitempty"`
	TargetID  int       `json:"target_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Room struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPrivate   bool      `json:"is_private"`
	CreatedBy   int       `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

func (db *DB) GetDMMessages(userID, targetID int) ([]Message, error) {
	query := `
		SELECT m.id, m.content, m.user_id, u.username, m.target_id, m.created_at
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE (m.user_id = $1 AND m.target_id = $2)
		OR (m.user_id = $2 AND m.target_id = $1)
		ORDER BY m.created_at ASC
	`
	rows, err := db.Query(query, userID, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.Content, &m.UserID, &m.Username, &m.TargetID, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (db *DB) GetPublicRooms() ([]Room, error) {
	query := `
		SELECT id, name, COALESCE(description, '') as description, is_private, COALESCE(creator_id, 0) as creator_id, created_at
		FROM rooms
		WHERE is_private = false
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var r Room
		err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsPrivate, &r.CreatedBy, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	return rooms, nil
}

func (db *DB) CreateRoom(name, description string, isPrivate bool, userID int) (int, error) {
	query := `
		INSERT INTO rooms (name, description, is_private, creator_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var roomID int
	err := db.QueryRow(query, name, description, isPrivate, userID).Scan(&roomID)
	if err != nil {
		return 0, err
	}
	return roomID, nil
}

func (db *DB) GetRoomMessages(roomID string) ([]Message, error) {
	query := `
		SELECT m.id, m.content, m.user_id, u.username, m.room_id, m.created_at
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.room_id = $1
		ORDER BY m.created_at ASC
	`
	rows, err := db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.Content, &m.UserID, &m.Username, &m.RoomID, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (db *DB) HasRoomAccess(userID int, roomID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM rooms
			WHERE id = $1
			AND (is_private = false OR created_by = $2)
		)
	`
	var hasAccess bool
	err := db.QueryRow(query, roomID, userID).Scan(&hasAccess)
	if err != nil {
		return false, err
	}
	return hasAccess, nil
}

func (db *DB) SendRoomMessage(roomID string, userID int, content string) (*Message, error) {
	query := `
		INSERT INTO messages (content, user_id, room_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	var message Message
	err := db.QueryRow(query, content, userID, roomID).Scan(&message.ID, &message.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Récupérer les informations complètes du message
	query = `
		SELECT m.content, m.user_id, u.username, m.room_id
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.id = $1
	`
	err = db.QueryRow(query, message.ID).Scan(&message.Content, &message.UserID, &message.Username, &message.RoomID)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (db *DB) SendDMMessage(userID, targetID int, content string) (*Message, error) {
	query := `
		INSERT INTO messages (content, user_id, target_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	var message Message
	err := db.QueryRow(query, content, userID, targetID).Scan(&message.ID, &message.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Récupérer les informations complètes du message
	query = `
		SELECT m.content, m.user_id, u.username, m.target_id
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.id = $1
	`
	err = db.QueryRow(query, message.ID).Scan(&message.Content, &message.UserID, &message.Username, &message.TargetID)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
