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
func FormatTags(tags []string) []string {
	encountered := map[string]bool{}
	formatted := make([]string, 0, len(tags))

	for _, tag := range tags {
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
			formatted = append(formatted, tag)
			encountered[tag] = true
		}
	}

	return formatted
}

// ToSearchTags transforms tags from user inputs into searching tags.
// dog walking -> dog, walking (two words)
func ToSearchTags(words string) []string {
	splitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}
	tags := strings.FieldsFunc(strings.ToLower(words), splitFn)
	return FormatTags(tags)
}

// ToTagFields converts tags into TagFields.
func ToTagFields(tags []string) []*types.TagField {
	tagFields := make([]*types.TagField, 0, len(tags))
	for _, tagName := range tags {
		tagField := &types.TagField{
			Name:      tagName,
			CreatedAt: time.Now(),
		}
		tagFields = append(tagFields, tagField)
	}
	return tagFields
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

func TagFieldToNames(tags []*types.TagField) []string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	return names
}

func TagToNames(tags []*types.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	return names
}
