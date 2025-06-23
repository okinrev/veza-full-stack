//file: backend/services/offer_service.go

package services

type OfferService interface {
    CreateOffer(listingID, fromUserID, proposedProductID int, message *string) (*models.Offer, error)
    AcceptOffer(offerID, userID int) error
    RejectOffer(offerID, userID int) error
    GetUserOffers(userID int, offerType string, page, limit int) ([]models.Offer, int, error)
}