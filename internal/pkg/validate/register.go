package validate

import (
	"strings"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

// ValidateEntity validates
// EntityName, Offers and Wants
func ValidateEntity(b *types.EntityData) []string {
	errs := []string{}
	if b.EntityName == "" {
		errs = append(errs, "Entity name is missing.")
	} else if len(b.EntityName) > 100 {
		errs = append(errs, "Entity Name cannot exceed 100 characters.")
	}
	if b.Website != "" && !strings.HasPrefix(b.Website, "http://") && !strings.HasPrefix(b.Website, "https://") {
		errs = append(errs, "Website URL should start with http:// or https://.")
	} else if len(b.Website) > 100 {
		errs = append(errs, "Website URL cannot exceed 100 characters.")
	}
	errs = append(errs, validateTagsLimit(b)...)
	return errs
}
