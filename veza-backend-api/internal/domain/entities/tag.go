package entities

import (
	"time"
)

// TagType représente le type de tag
type TagType string

const (
	TagTypeCategory   TagType = "category"
	TagTypeGenre      TagType = "genre"
	TagTypeMood       TagType = "mood"
	TagTypeInstrument TagType = "instrument"
	TagTypeCustom     TagType = "custom"
)

// Tag représente un tag dans le système
type Tag struct {
	ID          uint       `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Type        TagType    `json:"type" db:"type"`
	Color       string     `json:"color" db:"color"`
	Icon        string     `json:"icon" db:"icon"`
	UsageCount  int        `json:"usage_count" db:"usage_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TagCreateRequest représente la requête de création d'un tag
type TagCreateRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=50"`
	Description string  `json:"description" validate:"omitempty,max=200"`
	Type        TagType `json:"type" validate:"required"`
	Color       string  `json:"color" validate:"omitempty,hexcolor"`
	Icon        string  `json:"icon" validate:"omitempty"`
}

// TagUpdateRequest représente la requête de mise à jour d'un tag
type TagUpdateRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=1,max=50"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=200"`
	Type        *TagType `json:"type,omitempty"`
	Color       *string  `json:"color,omitempty" validate:"omitempty,hexcolor"`
	Icon        *string  `json:"icon,omitempty"`
}

// TagFilters représente les filtres pour la recherche de tags
type TagFilters struct {
	Type       TagType `json:"type,omitempty"`
	SearchTerm string  `json:"search_term,omitempty"`
	MinUsage   *int    `json:"min_usage,omitempty"`
}
