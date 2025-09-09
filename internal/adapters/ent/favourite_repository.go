package entadapter

import (
	"context"

	"github.com/SokratisChaimanas/platform-go-challenge/ent"
	"github.com/SokratisChaimanas/platform-go-challenge/ent/favourite"
	"github.com/SokratisChaimanas/platform-go-challenge/internal/domain"
	"github.com/SokratisChaimanas/platform-go-challenge/internal/ports"
	"github.com/google/uuid"
)

// Compile time safety for ports.FavouriteRepository implementation.
var _ ports.FavouriteRepository = (*FavouriteRepo)(nil)

// FavouriteRepo implements ports.FavouriteRepository using Ent.
type FavouriteRepo struct {
	client *ent.Client
}

func NewFavouriteRepo(client *ent.Client) *FavouriteRepo {
	return &FavouriteRepo{client: client}
}

// Create inserts a new favourite. Duplicate entries map to ErrFavouriteAlreadyExists.
func (favouriteRepo *FavouriteRepo) Create(ctx context.Context, favouriteToCreate *domain.Favourite) error {
	_, err := favouriteRepo.client.Favourite.
		Create().
		SetUserID(favouriteToCreate.UserID).
		SetAssetID(favouriteToCreate.AssetID).
		SetCreatedAt(favouriteToCreate.CreatedAt).
		Save(ctx)

	return err
}

// Delete removes a favourite by (userID, assetID). Missing rows map to ErrFavouriteNotFound.
func (favouriteRepo *FavouriteRepo) Delete(ctx context.Context, userID, assetID uuid.UUID) error {
	f, err := favouriteRepo.client.Favourite.
		Delete().
		Where(
			favourite.UserID(userID),
			favourite.AssetID(assetID),
		).
		Exec(ctx)

	if err != nil {
		return err
	}

	if f == 0 {
		return domain.ErrFavouriteNotFound
	}

	return nil
}

// Exists checks if a favourite already exists for (userID, assetID).
func (favouriteRepo *FavouriteRepo) Exists(ctx context.Context, userID, assetID uuid.UUID) (bool, error) {
	return favouriteRepo.client.Favourite.
		Query().
		Where(
			favourite.UserID(userID),
			favourite.AssetID(assetID),
		).
		Exist(ctx)
}

// ListAssetsFavouritedByUser returns the *assets* a user has favourited,
// ordered by favourite.created_at DESC (newest first), with pagination.
// Uses eager loading (WithAsset) to avoid N+1.
func (favouriteRepo *FavouriteRepo) ListAssetsFavouritedByUser(
	ctx context.Context,
	userID uuid.UUID,
	opt ports.ListOptions,
) ([]domain.Asset, error) {

	query := favouriteRepo.client.Favourite.
		Query().
		Where(favourite.UserID(userID)).
		Order(ent.Desc(favourite.FieldCreatedAt)).
		WithAsset()

	if opt.Limit >= 0 {
		query = query.Limit(opt.Limit)
	}
	if opt.Offset >= 0 {
		query = query.Offset(opt.Offset)
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	assets := make([]domain.Asset, 0, len(rows))
	for _, f := range rows {
		a := f.Edges.Asset
		if a == nil {
			continue
		}
		assets = append(assets, domain.Asset{
			ID:          a.ID,
			Type:        domain.AssetType(a.AssetType),
			Description: a.Description,
			Payload:     a.Payload,
		})
	}
	return assets, nil
}
