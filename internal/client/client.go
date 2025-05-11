package client

import (
	"time"

	"github.com/google/uuid"
)

// Note that one client == one unique API token, where client_id = api_token

type Client struct {
	ID         uuid.UUID
	Capacity   int
	RatePerSec int
	Tokens     int
	LastRefill time.Time
}

type ClientRepo interface {
	Create(c *Client) error
	Delete(id uuid.UUID) error
	Get(id uuid.UUID) (*Client, error)
	List() ([]*Client, error)
	UpdateTokens(id uuid.UUID, tokens int, lastRefill time.Time) error
}
