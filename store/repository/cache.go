package repository

import (
	"golang.org/x/net/context"
	"main/store/repository/models"
)

type Cache interface {
	Get(context.Context, string) (models.Cost, error)
	Set(context.Context, models.Cost) error
}
