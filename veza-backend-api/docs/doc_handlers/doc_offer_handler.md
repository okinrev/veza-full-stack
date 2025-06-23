# OfferHandler Documentation

This file refer to : backend/internal/handlers/offer.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, time, strconv, strings, net/http, database/sql
```

## Struct
```go
type OfferHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewOfferHandler(db *database.DB) *OfferHandler
```

## Types/Models
```go
type OfferResponse struct {
    ID, ListingID, FromUserID, ProposedProductID int
    FromUsername, ProposedProductName, Status string
    FromUserAvatar *string
    Message, CounterOffer, ExpiresAt, ViewedAt *string
    CreatedAt, UpdatedAt string
    ListingTitle, ListingDescription *string
    ResponseDeadline *string
}

type CreateOfferRequest struct {
    ProposedProductID int `binding:"required"`  // user_products.id
    Message *string
    ExpiresIn int  // Hours until expiration
}

type UpdateOfferRequest struct {
    Status *string        // "accepted", "rejected", "withdrawn", "countered"
    CounterOffer *string
    Message *string
}

type OfferStats struct {
    TotalOffers int
    OffersByStatus map[string]int
    AcceptanceRate float64
    AverageResponseTime, RecentOffers string  // Not implemented
}
```

## Methods

### Offer Management
```go
func (h *OfferHandler) CreateOffer(c *gin.Context)
func (h *OfferHandler) GetOffer(c *gin.Context)
func (h *OfferHandler) UpdateOffer(c *gin.Context)
func (h *OfferHandler) GetUserOffers(c *gin.Context)
func (h *OfferHandler) GetOffersForListing(c *gin.Context)
func (h *OfferHandler) GetOfferStats(c *gin.Context)
```

### CreateOffer
- **Auth**: common.GetUserIDFromContext required
- **Params**: listing_id (URL param)
- **Validation**: Listing exists/open, product ownership, no existing pending offer
- **Process**: Creates offer with optional expiration, status 'pending'

### GetOffer
- **Auth**: Required (offer creator OR listing owner)
- **Process**: Retrieves offer with auto-mark as viewed for listing owner
- **Database**: Complex JOIN for offer details

### UpdateOffer
- **Auth**: Required with role-based actions
- **Actions**:
  - Accept/Reject: Listing owner only
  - Withdraw: Offer creator only
  - Counter: Listing owner only
- **Transactions**: Acceptance triggers listing close + other offer rejections

### GetUserOffers
- **Auth**: Required
- **Params**: type (sent/received/all), status, page, limit
- **Process**: Dynamic query based on user relationship to offers

### GetOffersForListing
- **Auth**: Required (listing owner only)
- **Process**: All offers for specific listing

### GetOfferStats
- **Auth**: Required
- **Process**: Statistics for user's offers (sent/received)

## Helper Methods
```go
func (h *OfferHandler) getOfferByID(offerID, currentUserID int, isOfferCreator bool) (*OfferResponse, error)
```

## Status Workflow
```
pending → accepted/rejected/withdrawn/countered
countered → accepted/rejected/withdrawn
accepted → final (triggers listing closure)
```

## Transaction Logic (Accept)
1. Update offer status to 'accepted'
2. Update listing status to 'closed'
3. Reject all other pending/countered offers on same listing
4. Commit or rollback all changes

## Route Mapping Expectations
- `POST /listings/:listing_id/offers` → CreateOffer
- `GET /offers/:id` → GetOffer
- `PUT /offers/:id` → UpdateOffer
- `GET /offers` → GetUserOffers
- `GET /listings/:listing_id/offers` → GetOffersForListing
- `GET /offers/stats` → GetOfferStats

## Middleware Dependencies
- Authentication middleware
- JSON binding for POST/PUT requests

## Database Tables
- offers (main offer data)
- listings (listing details, status updates)
- users (user details for display)
- user_products (product ownership validation)
- products (product details for display)

## Key Relationships
- `offers.listing_id` → `listings.id`
- `offers.from_user_id` → `users.id`
- `offers.proposed_product_id` → `user_products.id`
- `listings.user_id` → `users.id` (listing owner)

## Authorization Matrix
| Action | Creator | Listing Owner | Other |
|--------|---------|---------------|-------|
| View | ✓ | ✓ | ✗ |
| Accept/Reject | ✗ | ✓ | ✗ |
| Withdraw | ✓ | ✗ | ✗ |
| Counter | ✗ | ✓ | ✗ |

## Error Patterns
- 401: User not authenticated
- 403: Not authorized (role-based actions)
- 404: Offer/listing not found
- 409: Existing pending offer, invalid status transition
- 500: Database/transaction errors