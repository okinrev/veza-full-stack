// internal/services/chat_service.go

package services

type ChatService interface {
    SendDirectMessage(fromUserID, toUserID int, content string) (*models.Message, error)
    GetDirectMessages(userID, otherUserID int, page, limit int) ([]models.Message, error)
    GetConversations(userID int, limit int) ([]models.ConversationSummary, error)
    MarkAsRead(fromUserID, toUserID int) error
}