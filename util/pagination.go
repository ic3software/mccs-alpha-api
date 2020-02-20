package util

import "math"

func GetNumberOfPages(numberOfResults int, sizeSize int) int {
	pages := int(math.Ceil(float64(numberOfResults) / float64(sizeSize)))
	if numberOfResults == 0 {
		return 1
	}
	return pages
}
