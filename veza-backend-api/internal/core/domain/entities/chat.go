package entities

import (
	"errors"
	"time"
)

// RoomType définit les types de salles de chat
type RoomType string

const (
	RoomTypePublic  RoomType = "public"
	RoomTypePrivate RoomType = "private"
	RoomTypePremium RoomType = "premium"
	RoomTypeDirect  RoomType = "direct"  // Messages directs
	RoomTypeGroup   RoomType = "group"   // Groupe privé
	RoomTypeStream  RoomType = "stream"  // Chat de stream
	RoomTypeSupport RoomType = "support" // Support client
)

// RoomPrivacy définit les niveaux de confidentialité des salles
type RoomPrivacy string

const (
	RoomPrivacyPublic     RoomPrivacy = "public"
	RoomPrivacyPrivate    RoomPrivacy = "private"
	RoomPrivacyInviteOnly RoomPrivacy = "invite_only"
	RoomPrivacyPassword   RoomPrivacy = "password"
)

// RoomStatus définit le statut d'une salle
type RoomStatus string

const (
	RoomStatusActive    RoomStatus = "active"
	RoomStatusInactive  RoomStatus = "inactive"
	RoomStatusArchived  RoomStatus = "archived"
	RoomStatusSuspended RoomStatus = "suspended"
	RoomStatusDeleted   RoomStatus = "deleted"
)

// MessageType définit les types de messages
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeFile     MessageType = "file"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeSystem   MessageType = "system"
	MessageTypeCommand  MessageType = "command"
	MessageTypeReaction MessageType = "reaction"
)

// MessageStatus définit le statut des messages
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusEdited    MessageStatus = "edited"
	MessageStatusDeleted   MessageStatus = "deleted"
	MessageStatusFlagged   MessageStatus = "flagged"
)

// Room représente une salle de chat
type Room struct {
	ID          int64       `json:"id" db:"id"`
	UUID        string      `json:"uuid" db:"uuid"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Type        RoomType    `json:"type" db:"type"`
	Privacy     RoomPrivacy `json:"privacy" db:"privacy"`
	Status      RoomStatus  `json:"status" db:"status"`

	// Paramètres de la salle
	IsActive     bool     `json:"is_active" db:"is_active"`
	MaxMembers   int      `json:"max_members" db:"max_members"`
	RequiredRole UserRole `json:"required_role" db:"required_role"`
	Password     string   `json:"-" db:"password_hash"`
	Topic        string   `json:"topic" db:"topic"`

	// Propriétaire et modération
	OwnerID      int64   `json:"owner_id" db:"owner_id"`
	ModeratorIDs []int64 `json:"moderator_ids" db:"moderator_ids"`

	// Statistiques
	MemberCount   int   `json:"member_count" db:"member_count"`
	MessageCount  int64 `json:"message_count" db:"message_count"`
	LastMessageID int64 `json:"last_message_id" db:"last_message_id"`

	// Métadonnées
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Relations (chargées séparément)
	Owner       *User    `json:"owner,omitempty"`
	LastMessage *Message `json:"last_message,omitempty"`
	Members     []*User  `json:"members,omitempty"`
}

// ChatRoom est un alias pour Room pour maintenir la cohérence avec les repositories
type ChatRoom = Room

// ChatMessage est un alias pour Message pour maintenir la cohérence avec les repositories
type ChatMessage = Message

// RoomRole définit les rôles dans une salle
type RoomRole string

const (
	RoomRoleMember    RoomRole = "member"
	RoomRoleModerator RoomRole = "moderator"
	RoomRoleAdmin     RoomRole = "admin"
	RoomRoleOwner     RoomRole = "owner"
)

// InvitationStatus définit le statut d'une invitation
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusRevoked  InvitationStatus = "revoked"
)

// Message représente un message dans une salle
type Message struct {
	ID     int64  `json:"id" db:"id"`
	UUID   string `json:"uuid" db:"uuid"`
	RoomID int64  `json:"room_id" db:"room_id"`
	UserID int64  `json:"user_id" db:"user_id"`

	// Contenu du message
	Type    MessageType `json:"type" db:"type"`
	Content string      `json:"content" db:"content"`

	// Messages composés (images, fichiers, etc.)
	Attachments []MessageAttachment `json:"attachments,omitempty"`

	// Métadonnées
	Status   MessageStatus `json:"status" db:"status"`
	IsEdited bool          `json:"is_edited" db:"is_edited"`
	IsSystem bool          `json:"is_system" db:"is_system"`

	// Message parent (pour les réponses)
	ParentID    *int64 `json:"parent_id,omitempty" db:"parent_id"`
	ThreadCount int    `json:"thread_count" db:"thread_count"`

	// Modération
	IsFlagged   bool   `json:"is_flagged" db:"is_flagged"`
	FlagReason  string `json:"flag_reason,omitempty" db:"flag_reason"`
	ModeratedBy *int64 `json:"moderated_by,omitempty" db:"moderated_by"`

	// Réactions et mentions
	ReactionCount  int     `json:"reaction_count" db:"reaction_count"`
	MentionedUsers []int64 `json:"mentioned_users" db:"mentioned_users"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	EditedAt  *time.Time `json:"edited_at,omitempty" db:"edited_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Relations (chargées séparément)
	User          *User              `json:"user,omitempty"`
	Room          *Room              `json:"room,omitempty"`
	ParentMessage *Message           `json:"parent_message,omitempty"`
	Reactions     []*MessageReaction `json:"reactions,omitempty"`
}

// MessageAttachment représente une pièce jointe
type MessageAttachment struct {
	ID           int64  `json:"id" db:"id"`
	MessageID    int64  `json:"message_id" db:"message_id"`
	FileName     string `json:"file_name" db:"file_name"`
	FileSize     int64  `json:"file_size" db:"file_size"`
	MimeType     string `json:"mime_type" db:"mime_type"`
	URL          string `json:"url" db:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty" db:"thumbnail_url"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MessageReaction représente une réaction à un message
type MessageReaction struct {
	ID        int64  `json:"id" db:"id"`
	MessageID int64  `json:"message_id" db:"message_id"`
	UserID    int64  `json:"user_id" db:"user_id"`
	Emoji     string `json:"emoji" db:"emoji"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Relations
	User *User `json:"user,omitempty"`
}

// RoomMember représente un membre d'une salle
type RoomMember struct {
	ID     int64    `json:"id" db:"id"`
	RoomID int64    `json:"room_id" db:"room_id"`
	UserID int64    `json:"user_id" db:"user_id"`
	Role   UserRole `json:"role" db:"role"` // Rôle dans cette salle spécifique

	// Permissions spécifiques
	CanWrite    bool       `json:"can_write" db:"can_write"`
	CanModerate bool       `json:"can_moderate" db:"can_moderate"`
	IsMuted     bool       `json:"is_muted" db:"is_muted"`
	MutedUntil  *time.Time `json:"muted_until,omitempty" db:"muted_until"`

	// Dernière activité
	LastReadMessageID int64     `json:"last_read_message_id" db:"last_read_message_id"`
	LastSeenAt        time.Time `json:"last_seen_at" db:"last_seen_at"`

	// Métadonnées
	JoinedAt time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt   *time.Time `json:"left_at,omitempty" db:"left_at"`

	// Relations
	User *User `json:"user,omitempty"`
	Room *Room `json:"room,omitempty"`
}

// DirectConversation représente une conversation directe entre deux utilisateurs
type DirectConversation struct {
	ID      int64 `json:"id" db:"id"`
	User1ID int64 `json:"user1_id" db:"user1_id"`
	User2ID int64 `json:"user2_id" db:"user2_id"`

	// Dernière activité
	LastMessageID int64     `json:"last_message_id" db:"last_message_id"`
	LastActivity  time.Time `json:"last_activity" db:"last_activity"`

	// État de la conversation
	IsBlocked  bool   `json:"is_blocked" db:"is_blocked"`
	BlockedBy  *int64 `json:"blocked_by,omitempty" db:"blocked_by"`
	IsArchived bool   `json:"is_archived" db:"is_archived"`

	// Métadonnées
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Relations
	User1       *User    `json:"user1,omitempty"`
	User2       *User    `json:"user2,omitempty"`
	LastMessage *Message `json:"last_message,omitempty"`
}

// NewRoom crée une nouvelle salle de chat
func NewRoom(name, description string, roomType RoomType, ownerID int64) (*Room, error) {
	room := &Room{
		Name:         name,
		Description:  description,
		Type:         roomType,
		OwnerID:      ownerID,
		IsActive:     true,
		MaxMembers:   1000, // Par défaut
		RequiredRole: RoleUser,
		MemberCount:  1, // Le propriétaire
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := room.Validate(); err != nil {
		return nil, err
	}

	return room, nil
}

// Validate vérifie que la salle est valide
func (r *Room) Validate() error {
	if r.Name == "" {
		return errors.New("room name is required")
	}

	if len(r.Name) < 2 {
		return errors.New("room name must be at least 2 characters long")
	}

	if len(r.Name) > 100 {
		return errors.New("room name must be less than 100 characters long")
	}

	if err := r.ValidateType(); err != nil {
		return err
	}

	if r.MaxMembers < 2 {
		return errors.New("room must allow at least 2 members")
	}

	return nil
}

// ValidateType vérifie que le type de salle est valide
func (r *Room) ValidateType() error {
	switch r.Type {
	case RoomTypePublic, RoomTypePrivate, RoomTypePremium, RoomTypeDirect, RoomTypeGroup, RoomTypeStream, RoomTypeSupport:
		return nil
	default:
		return errors.New("invalid room type")
	}
}

// CanUserJoin vérifie si un utilisateur peut rejoindre la salle
func (r *Room) CanUserJoin(user *User) bool {
	if !r.IsActive {
		return false
	}

	// Vérifier le rôle requis
	if !user.HasAnyRole(r.RequiredRole, RoleAdmin, RoleSuperAdmin) {
		return false
	}

	// Vérifier la capacité
	if r.MemberCount >= r.MaxMembers {
		return false
	}

	return true
}

// IsOwner vérifie si un utilisateur est propriétaire de la salle
func (r *Room) IsOwner(userID int64) bool {
	return r.OwnerID == userID
}

// IsModerator vérifie si un utilisateur est modérateur de la salle
func (r *Room) IsModerator(userID int64) bool {
	if r.IsOwner(userID) {
		return true
	}

	for _, modID := range r.ModeratorIDs {
		if modID == userID {
			return true
		}
	}

	return false
}

// NewMessage crée un nouveau message
func NewMessage(roomID, userID int64, messageType MessageType, content string) (*Message, error) {
	message := &Message{
		RoomID:        roomID,
		UserID:        userID,
		Type:          messageType,
		Content:       content,
		Status:        MessageStatusSent,
		IsEdited:      false,
		IsSystem:      false,
		IsFlagged:     false,
		ReactionCount: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := message.Validate(); err != nil {
		return nil, err
	}

	return message, nil
}

// Validate vérifie que le message est valide
func (m *Message) Validate() error {
	if m.RoomID == 0 {
		return errors.New("room ID is required")
	}

	if m.UserID == 0 {
		return errors.New("user ID is required")
	}

	if err := m.ValidateType(); err != nil {
		return err
	}

	if m.Type == MessageTypeText && m.Content == "" {
		return errors.New("text message content is required")
	}

	if len(m.Content) > 10000 {
		return errors.New("message content too long (max 10000 characters)")
	}

	return nil
}

// ValidateType vérifie que le type de message est valide
func (m *Message) ValidateType() error {
	switch m.Type {
	case MessageTypeText, MessageTypeImage, MessageTypeFile, MessageTypeAudio, MessageTypeVideo, MessageTypeSystem, MessageTypeCommand, MessageTypeReaction:
		return nil
	default:
		return errors.New("invalid message type")
	}
}

// CanBeEditedBy vérifie si un message peut être édité par un utilisateur
func (m *Message) CanBeEditedBy(userID int64) bool {
	// Seul l'auteur peut éditer son message
	if m.UserID != userID {
		return false
	}

	// Les messages système ne peuvent pas être édités
	if m.IsSystem {
		return false
	}

	// Ne peut pas éditer un message supprimé
	if m.DeletedAt != nil {
		return false
	}

	// Limite de temps pour l'édition (15 minutes)
	if time.Since(m.CreatedAt) > 15*time.Minute {
		return false
	}

	return true
}

// CanBeDeletedBy vérifie si un message peut être supprimé par un utilisateur
func (m *Message) CanBeDeletedBy(userID int64, userRole UserRole, isRoomModerator bool) bool {
	// L'auteur peut toujours supprimer son message
	if m.UserID == userID {
		return true
	}

	// Les modérateurs de la salle peuvent supprimer les messages
	if isRoomModerator {
		return true
	}

	// Les admins peuvent supprimer n'importe quel message
	if userRole == RoleAdmin || userRole == RoleSuperAdmin {
		return true
	}

	return false
}

// Edit édite le contenu d'un message
func (m *Message) Edit(newContent string, userID int64) error {
	if !m.CanBeEditedBy(userID) {
		return errors.New("message cannot be edited by this user")
	}

	if newContent == "" {
		return errors.New("edited content cannot be empty")
	}

	if len(newContent) > 10000 {
		return errors.New("edited content too long (max 10000 characters)")
	}

	m.Content = newContent
	m.IsEdited = true
	m.Status = MessageStatusEdited
	now := time.Now()
	m.EditedAt = &now
	m.UpdatedAt = now

	return nil
}

// SoftDelete marque un message comme supprimé
func (m *Message) SoftDelete() {
	now := time.Now()
	m.DeletedAt = &now
	m.Status = MessageStatusDeleted
	m.UpdatedAt = now
}

// Flag marque un message comme signalé
func (m *Message) Flag(reason string, moderatorID int64) {
	m.IsFlagged = true
	m.FlagReason = reason
	m.ModeratedBy = &moderatorID
	m.UpdatedAt = time.Now()
}

// NewDirectConversation crée une nouvelle conversation directe
func NewDirectConversation(user1ID, user2ID int64) (*DirectConversation, error) {
	if user1ID == user2ID {
		return nil, errors.New("cannot create conversation with yourself")
	}

	// S'assurer que user1ID < user2ID pour la cohérence
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	conversation := &DirectConversation{
		User1ID:      user1ID,
		User2ID:      user2ID,
		LastActivity: time.Now(),
		IsBlocked:    false,
		IsArchived:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return conversation, nil
}

// GetOtherUserID retourne l'ID de l'autre utilisateur dans la conversation
func (dc *DirectConversation) GetOtherUserID(currentUserID int64) int64 {
	if dc.User1ID == currentUserID {
		return dc.User2ID
	}
	return dc.User1ID
}

// IsParticipant vérifie si un utilisateur fait partie de la conversation
func (dc *DirectConversation) IsParticipant(userID int64) bool {
	return dc.User1ID == userID || dc.User2ID == userID
}

// Block bloque la conversation
func (dc *DirectConversation) Block(blockedBy int64) error {
	if !dc.IsParticipant(blockedBy) {
		return errors.New("user is not a participant in this conversation")
	}

	dc.IsBlocked = true
	dc.BlockedBy = &blockedBy
	dc.UpdatedAt = time.Now()

	return nil
}

// Unblock débloque la conversation
func (dc *DirectConversation) Unblock() {
	dc.IsBlocked = false
	dc.BlockedBy = nil
	dc.UpdatedAt = time.Now()
}
