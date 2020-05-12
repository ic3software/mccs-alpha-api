package util

import (
	"fmt"
	"reflect"

	"gopkg.in/oleiade/reflections.v1"
)

var defaultFieldsToSkip = []string{
	"CurrentLoginIP",
	"Password",
	"LastLoginIP",
	"UpdatedAt",
}

// CheckFieldDiff checks what fields have been changed.
func CheckFieldDiff(oldStruct interface{}, newStruct interface{}, fieldsToSkip ...string) []string {
	modifiedFields := []string{}

	structItems, _ := reflections.Items(oldStruct)
	skipMap := sliceToMap(append(fieldsToSkip, defaultFieldsToSkip...))

	for field, origin := range structItems {
		if _, ok := skipMap[field]; ok {
			continue
		}
		update, _ := reflections.GetField(newStruct, field)
		if !reflect.DeepEqual(origin, update) {
			fieldKind, _ := reflections.GetFieldKind(oldStruct, field)
			switch fieldKind {
			case reflect.String:
				modifiedFields = append(modifiedFields, handleString(field, origin, update))
			case reflect.Int:
			case reflect.Int32:
			case reflect.Int64:
				modifiedFields = append(modifiedFields, handleInt(field, origin, update))
			case reflect.Float32:
			case reflect.Float64:
				modifiedFields = append(modifiedFields, handleFloat(field, origin, update))
			case reflect.Bool:
				modifiedFields = append(modifiedFields, handleBool(field, origin, update))
			case reflect.Ptr:
				modifiedFields = append(modifiedFields, handlePtr(field, origin, update))
			case reflect.Slice:
				modifiedFields = append(modifiedFields, handleSlice(field, origin, update))
			}
		}
	}

	return modifiedFields
}

func handleString(field string, origin interface{}, update interface{}) string {
	return fmt.Sprintf("%s: %s -> %s", field, origin, update)
}

func handleInt(field string, origin interface{}, update interface{}) string {
	return fmt.Sprintf("%s: %d -> %d", field, origin, update)
}

func handleFloat(field string, origin interface{}, update interface{}) string {
	return fmt.Sprintf("%s: %.2f -> %.2f", field, origin, update)
}

func handleBool(field string, origin interface{}, update interface{}) string {
	return fmt.Sprintf("%s: %t -> %t", field, origin, update)
}

func handlePtr(field string, origin interface{}, update interface{}) string {
	intPtr, ok := origin.(*int)
	if ok {
		updateIntPtr, _ := update.(*int)
		if intPtr == nil {
			return handleInt(field, 0, *updateIntPtr)
		} else {
			return handleInt(field, *intPtr, *updateIntPtr)
		}
	}
	floatPtr, ok := origin.(*float64)
	if ok {
		updateFloatPtr, _ := update.(*float64)
		if floatPtr == nil {
			return handleFloat(field, 0, *updateFloatPtr)
		} else {
			return handleFloat(field, *floatPtr, *updateFloatPtr)
		}
	}
	boolPtr, ok := origin.(*bool)
	if ok {
		updateBoolPtr, _ := update.(*bool)
		if boolPtr == nil {
			return handleBool(field, false, *updateBoolPtr)
		} else {
			return handleBool(field, *boolPtr, *updateBoolPtr)
		}
	}
	return ""
}

func handleSlice(field string, origin interface{}, update interface{}) string {
	return fmt.Sprintf("%s: %q -> %q", field, origin, update)
}

func sliceToMap(elements []string) map[string]bool {
	elementMap := make(map[string]bool)
	for _, e := range elements {
		elementMap[e] = true
	}
	return elementMap
}
