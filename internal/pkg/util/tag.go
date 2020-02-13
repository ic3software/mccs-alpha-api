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

// TagDifference finds out the new added tags.
func TagDifference(new, old []string) ([]string, []string) {
	encountered := map[string]int{}
	added := []string{}
	removed := []string{}
	for _, tag := range old {
		if _, ok := encountered[tag]; !ok {
			encountered[tag]++
		}
	}
	for _, tag := range new {
		encountered[tag]--
	}
	for name, flag := range encountered {
		if flag == -1 {
			added = append(added, name)
		}
		if flag == 1 {
			removed = append(removed, name)
		}
	}
	return added, removed
}

// GetTagNames gets tag name from TagField.
func GetTagNames(tags []*types.TagField) []string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	return names
}
