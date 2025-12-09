package utils

import "strings"

// check if atleast one of the product categories is in  request categories
func CheckCategory(productCategories []string, requestCateogies []string) bool {
	for _, c := range productCategories {
		for _, rc := range requestCateogies {
			if c == rc {
				return true
			}
		}
	}
	return false
}

func FilterName(input string, name string) bool {
	return strings.Contains(name, input)
}
