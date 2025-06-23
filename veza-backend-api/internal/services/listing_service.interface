//file: backend/services/listing_service.go

package services

type ListingService interface {
    CreateListing(userID int, req CreateListingRequest) (*models.Listing, error)
    GetListings(page, limit int, status string) ([]models.Listing, int, error)
    UpdateListing(listingID, userID int, req UpdateListingRequest) (*models.Listing, error)
    DeleteListing(listingID, userID int) error
}