package balancer

import (
	"github.com/zahartd/load_balancer/internal/models"
)

type Algorithm interface {
	Next(backends []*models.Backend) *models.Backend
}
