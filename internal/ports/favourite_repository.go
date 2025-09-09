package ports

import (
	"context"

	"github.com/SokratisChaimanas/platform-go-challenge/internal/domain"
	"github.com/google/uuid"
)

// ListOptions controls pagination when listing favourites.
type ListOptions struct {
	Limit  int
	Offset int
}

// FavouriteRepository stores and queries favourites.
type FavouriteRepository interface {
	// Create inserts a favourite. Duplicate should return domain.ErrFavouriteAlreadyExists.
	Create(ctx context.Context, f *domain.Favourite) error

	// Delete removes a favourite. Missing should return domain.ErrFavouriteNotFound.
	Delete(ctx context.Context, userID, assetID uuid.UUID) error

	// Exists checks whether (user, asset) is already favourited.
	Exists(ctx context.Context, userID, assetID uuid.UUID) (bool, error)

	// ListAssetsFavouritedByUser ListByUser returns a user's favourites (paged via ports.ListOptions).
	ListAssetsFavouritedByUser(ctx context.Context, userID uuid.UUID, opt ListOptions) ([]domain.Asset, error)
}
