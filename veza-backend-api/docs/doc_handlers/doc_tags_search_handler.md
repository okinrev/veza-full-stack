# TagsSearchHandler Documentation

This file refer to : backend/internal/handlers/tags_search.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, pq, sync, time, fmt, strconv, strings
```

## Struct
```go
type TagsSearchHandler struct {
    db    *database.DB
    cache *SuggestionCache
}
```

## Constructor
```go
func NewTagsSearchHandler(db *database.DB) *TagsSearchHandler
```

## Types/Models
```go
type GlobalSearchResult struct {
    Users []UserSearchResult
    Tracks []TrackSearchResult
    SharedResources []SharedResourceSearchResult
    TotalResults int
    Query string
}

type UserSearchResult struct {
    ID int
    Username, Email string
    FirstName, LastName, Avatar *string
}

type TrackSearchResult struct {
    ID, UploaderID int
    Title, Artist, Filename, UploaderName, CreatedAt, StreamURL string
    Tags []string
}

type SharedResourceSearchResult struct {
    ID, UploaderID int
    Title, Filename, Type, UploaderUsername, UploadedAt, DownloadURL string
    Description *string
    Tags []string
}

type AutocompleteResult struct {
    Tags, Artists, Users, Types []string
}

type SuggestionResponse struct {
    Type, Value string
    Count int
}

type ContextualSuggestion struct {
    Value, Type string
    Score float64
    Context map[string]string
    Usage int
    Trending bool
    Related []string
}
```

## Caching System
```go
type SuggestionCache struct {
    mu    sync.RWMutex
    cache map[string]CacheEntry
}

type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
}
```

## Methods

### Global Search
```go
func (h *TagsSearchHandler) GlobalSearch(c *gin.Context)
func (h *TagsSearchHandler) AdvancedSearch(c *gin.Context)
```
- **Auth**: None for global, required for advanced
- **Params**: q (query), type, tag, limit
- **Process**: Cross-entity search (users, tracks, resources)

### Tag Operations
```go
func (h *TagsSearchHandler) GetAllTags(c *gin.Context)
func (h *TagsSearchHandler) SearchTags(c *gin.Context)
func (h *TagsSearchHandler) GetTrendingTags(c *gin.Context)
```
- **Process**: Tag aggregation from tracks and shared_resources
- **Features**: Usage counting, trend analysis

### Suggestions
```go
func (h *TagsSearchHandler) GetAutocomplete(c *gin.Context)
func (h *TagsSearchHandler) GetSuggestions(c *gin.Context)
func (h *TagsSearchHandler) GetContextualSuggestions(c *gin.Context)
```
- **Features**: Multi-type autocomplete, context-aware suggestions
- **Caching**: 5-minute cache for contextual suggestions

### Search Helpers
```go
func (h *TagsSearchHandler) searchTracks(result *GlobalSearchResult, query, tag string, userID, limit int)
func (h *TagsSearchHandler) searchSharedResources(result *GlobalSearchResult, query, tag string, userID, limit int)
func (h *TagsSearchHandler) searchUsers(result *GlobalSearchResult, query string, userID, limit int)
```

### Context-Aware Suggestions
```go
func (h *TagsSearchHandler) getTrackContextSuggestions(query string, userID, limit int) []ContextualSuggestion
func (h *TagsSearchHandler) getResourceContextSuggestions(query string, userID, limit int) []ContextualSuggestion
func (h *TagsSearchHandler) getUserContextSuggestions(query string, userID, limit int) []ContextualSuggestion
func (h *TagsSearchHandler) getGlobalContextSuggestions(query string, userID, limit int) []ContextualSuggestion
```

### Cache Management
```go
func (h *TagsSearchHandler) ClearSuggestionCache(c *gin.Context)
```
- **Auth**: Admin required

## Route Mapping Expectations
- `GET /search` → GlobalSearch
- `GET /search/advanced` → AdvancedSearch
- `GET /tags` → GetAllTags
- `GET /tags/search` → SearchTags
- `GET /tags/trending` → GetTrendingTags
- `GET /autocomplete` → GetAutocomplete
- `GET /suggestions` → GetSuggestions
- `GET /suggestions/contextual` → GetContextualSuggestions
- `DELETE /admin/cache/suggestions` → ClearSuggestionCache

## Middleware Dependencies
- Authentication middleware for advanced search, admin operations
- Optional authentication for contextual suggestions

## Database Tables
- users (user search)
- tracks (with tags array)
- shared_resources (with tags array)

## Search Features
- **Global**: Cross-entity search with result aggregation
- **Advanced**: Filtered search by entity type
- **Tags**: Usage-based ranking, trend analysis
- **Autocomplete**: Multi-field suggestions (tags, artists, users, types)
- **Contextual**: Context-aware suggestions with scoring

## Caching Strategy
- **Duration**: 5-30 minutes depending on data type
- **Keys**: Query-based cache keys
- **Cleanup**: Background goroutine for expired entries
- **Admin**: Manual cache clearing capability

## Trend Analysis
- **Timeframes**: 7d, 30d, 90d
- **Calculation**: Recent usage percentage vs total usage
- **Minimum**: Requires >1 usage to qualify as trending

## Error Patterns
- 400: Invalid query parameters, minimum length requirements
- 401: Authentication required for certain operations
- 403: Admin access required for cache operations
- 500: Database errors, cache failures