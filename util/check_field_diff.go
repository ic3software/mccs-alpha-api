package util

import (
	"fmt"
	"reflect"
	"strconv"

	"gopkg.in/oleiade/reflections.v1"
)

var defaultFieldsToSkip = []string{"CurrentLoginIP", "Password", "LastLoginIP"}

// CheckDiff checks what fields have been changed.
// Only checks "String", "Int" and "Float64" types.
func CheckDiff(origin interface{}, updated interface{}, fieldsToSkip ...string) []string {
	modifiedFields := []string{}
	structItems, _ := reflections.Items(origin)
	skipMap := sliceToMap(append(fieldsToSkip, defaultFieldsToSkip...))

	for field, oldValue := range structItems {
		if _, ok := skipMap[field]; ok {
			continue
		}
		fieldKind, _ := reflections.GetFieldKind(origin, field)
		if fieldKind != reflect.String && fieldKind != reflect.Int && fieldKind != reflect.Float64 {
			continue
		}
		newValue, _ := reflections.GetField(updated, field)
		if newValue != oldValue {
			if fieldKind == reflect.Int {
				modifiedFields = append(modifiedFields, field+": "+strconv.Itoa(oldValue.(int))+" -> "+strconv.Itoa(newValue.(int)))
			} else if fieldKind == reflect.Float64 {
				modifiedFields = append(modifiedFields, field+": "+fmt.Sprintf("%.2f", oldValue.(float64))+" -> "+fmt.Sprintf("%.2f", newValue.(float64)))
			} else {
				modifiedFields = append(modifiedFields, field+": "+oldValue.(string)+" -> "+newValue.(string))
			}
		}
	}

	return modifiedFields
}

func sliceToMap(elements []string) map[string]bool {
	elementMap := make(map[string]bool)
	for _, e := range elements {
		elementMap[e] = true
	}
	return elementMap
}
