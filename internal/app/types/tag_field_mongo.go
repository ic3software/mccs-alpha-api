package types

import "time"

type TagField struct {
	Name      string    `json:"name,omitempty" bson:"name,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

func TagFieldToNames(tags []*TagField) []string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	return names
}

// ToTagFields converts tags into TagFields.
func ToTagFields(tags []string) []*TagField {
	tagFields := make([]*TagField, 0, len(tags))
	for _, tagName := range tags {
		tagField := &TagField{
			Name:      tagName,
			CreatedAt: time.Now(),
		}
		tagFields = append(tagFields, tagField)
	}
	return tagFields
}
