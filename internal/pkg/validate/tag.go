package validate

import (
	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/spf13/viper"
)

func SearchTag(query *types.SearchTagQuery) []error {
	errs := []error{}
	return errs
}

func validateTagsLimit(b *types.EntityData) []string {
	errorMessages := []string{}

	if len(b.Offers) == 0 {
		errorMessages = append(errorMessages, "Missing at least one valid tag for Products/Services Offered.")
	} else if len(b.Offers) > viper.GetInt("tags_limit") {
		errorMessages = append(errorMessages, "No more than "+viper.GetString("tags_limit")+" tags can be specified for Products/Services Offered.")
	}

	if len(b.Wants) == 0 {
		errorMessages = append(errorMessages, "Missing at least one valid tag for Products/Services Wanted.")
	} else if len(b.Wants) > viper.GetInt("tags_limit") {
		errorMessages = append(errorMessages, "No more than "+viper.GetString("tags_limit")+" tags can be specified for Products/Services Wanted.")
	}

	for _, offer := range b.Offers {
		if len(offer.Name) > 50 {
			errorMessages = append(errorMessages, "An Offer tag cannot exceed 50 characters.")
			break
		}
	}

	for _, want := range b.Wants {
		if len(want.Name) > 50 {
			errorMessages = append(errorMessages, "A Want tag cannot exceed 50 characters.")
			break
		}
	}

	return errorMessages
}
