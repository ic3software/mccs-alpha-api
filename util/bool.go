package util

func ToBool(boolean *bool) bool {
	if boolean == nil {
		return false
	}
	return *boolean
}
