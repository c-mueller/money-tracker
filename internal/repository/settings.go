package repository

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/ent"
	entsettings "icekalt.dev/money-tracker/ent/settings"
	"icekalt.dev/money-tracker/internal/domain"
)

type SettingsRepository struct {
	client *ent.Client
}

func NewSettingsRepository(client *ent.Client) *SettingsRepository {
	return &SettingsRepository{client: client}
}

func (r *SettingsRepository) Get(ctx context.Context, key string) (string, error) {
	s, err := r.client.Settings.Query().
		Where(entsettings.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", fmt.Errorf("%w: setting %s", domain.ErrNotFound, key)
		}
		return "", err
	}
	return s.Value, nil
}

func (r *SettingsRepository) Set(ctx context.Context, key, value string) error {
	existing, err := r.client.Settings.Query().
		Where(entsettings.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			_, err = r.client.Settings.Create().
				SetKey(key).
				SetValue(value).
				Save(ctx)
			return err
		}
		return err
	}
	_, err = existing.Update().SetValue(value).Save(ctx)
	return err
}
