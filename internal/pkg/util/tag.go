package util

import (
	"regexp"
	"strings"
	"time"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
)

var (
	specialCharRe *regexp.Regexp
	multiDashRe   *regexp.Regexp
	ltDashRe      *regexp.Regexp
	adminTagRe    *regexp.Regexp
)

func init() {
	specialCharRe = regexp.MustCompile("(&quot;)|([^a-zA-Z-]+)")
	multiDashRe = regexp.MustCompile("-+")
	ltDashRe = regexp.MustCompile("(^-+)|(-+$)")
	adminTagRe = regexp.MustCompile("[0-9]|(&quot;)|([^a-zA-Z ]+)")
}

// GetTags transforms tags from the user inputs into a standard format.
// dog walking -> dog-walking (one word)
func GetTags(tagArray []string) []*types.TagField {
	encountered := map[string]bool{}
	tags := make([]*types.TagField, 0, len(tagArray))

	for _, tag := range tagArray {
		tag = strings.ToLower(tag)
		tag = strings.Replace(tag, " ", "-", -1)
		tag = specialCharRe.ReplaceAllString(tag, "")
		tag = multiDashRe.ReplaceAllString(tag, "-")
		tag = ltDashRe.ReplaceAllString(tag, "")
		if len(tag) == 0 {
			continue
		}
		// remove duplicates
		if !encountered[tag] {
			tags = append(tags, &types.TagField{
				Name:      tag,
				CreatedAt: time.Now(),
			})
			encountered[tag] = true
		}
	}

	return tags
}
