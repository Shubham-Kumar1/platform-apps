package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eventhub/event-service/internal/models"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var u models.User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, email, name FROM users WHERE id = $1 AND is_active = true",
		userID,
	).Scan(&u.ID, &u.Email, &u.Name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	return &u, err
}

func (s *UserStore) GetOrganization(ctx context.Context, orgID string) (*models.Organization, error) {
	var o models.Organization
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, slug, plan_tier FROM organizations WHERE id = $1",
		orgID,
	).Scan(&o.ID, &o.Name, &o.Slug, &o.PlanTier)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization not found: %s", orgID)
	}
	return &o, err
}

func (s *UserStore) ValidateMembership(ctx context.Context, userID, orgID string) (string, error) {
	var role string
	err := s.db.QueryRowContext(ctx,
		"SELECT role FROM org_memberships WHERE user_id = $1 AND org_id = $2",
		userID, orgID,
	).Scan(&role)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("user %s is not a member of org %s", userID, orgID)
	}
	return role, err
}
