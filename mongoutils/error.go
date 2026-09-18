package mongoutils

import (
	"errors"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/errs"
	"go.mongodb.org/mongo-driver/mongo"
)

func ToDomainErr(err error, entity string) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errs.NewNotFound(err, fmt.Sprintf("%s not found", entity))
	} else if mongo.IsDuplicateKeyError(err) {
		return errs.NewAlreadyExists(err, fmt.Sprintf("%s already exists", entity))
	} else if mongo.IsTimeout(err) {
		return errs.NewTimeout(err)
	} else if mongo.IsNetworkError(err) {
		return errs.NewNetwork(err)
	}
	return err
}
