package validate

import (
	"strings"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

// ValidateBusiness validates
// BusinessName, Offers and Wants
func ValidateBusiness(b *types.BusinessData) []string {
	errs := []string{}
	if b.BusinessName == "" {
		errs = append(errs, "Business name is missing.")
	} else if len(b.BusinessName) > 100 {
		errs = append(errs, "Business Name cannot exceed 100 characters.")
	}
	if b.Website != "" && !strings.HasPrefix(b.Website, "http://") && !strings.HasPrefix(b.Website, "https://") {
		errs = append(errs, "Website URL should start with http:// or https://.")
	} else if len(b.Website) > 100 {
		errs = append(errs, "Website URL cannot exceed 100 characters.")
	}
	errs = append(errs, validateTagsLimit(b)...)
	return errs
}
