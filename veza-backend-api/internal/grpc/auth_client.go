package grpc

import (
	"context"
)

// AuthClient client gRPC pour le service d'authentification
type AuthClient struct {
	client AuthServiceClient
}

// NewAuthClient crée un nouveau client gRPC pour l'authentification
func NewAuthClient() *AuthClient {
	return &AuthClient{}
}

// ValidateToken valide un token JWT
func (a *AuthClient) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	// TODO: Implémenter l'appel gRPC réel
	isValid := req.Token != "" && len(req.Token) > 10
	
	response := &ValidateTokenResponse{
		Valid: isValid,
	}

	if !isValid {
		response.Error = "Invalid token format"
	}

	return response, nil
}
