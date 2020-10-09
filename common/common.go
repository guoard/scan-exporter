package common

import (
	"math/rand"
	"sort"
	"time"
)

// StringInSlice checks if a string appears in a slice.
func StringInSlice(s string, sl []string) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}
	return false
}

// CompareStringSlices checks if two slices are equal.
// It returns the number of different items.
func CompareStringSlices(sl1, sl2 []string) int {
	sort.Strings(sl1)
	sort.Strings(sl2)

	newports := []string{}
	missingports := []string{}

	for _, v := range sl2 {
		if !StringInSlice(v, sl1) {
			newports = append(newports, v)
		}
	}

	for _, v := range sl1 {
		if !StringInSlice(v, sl2) {
			missingports = append(missingports, v)
		}
	}

	return len(newports) + len(missingports)
}

// CompareStringSlices2 checks if two slices are equal.
// It returns the number of different items.
func CompareStringSlices2(sl1, sl2 []string) int {
	diffCounter := 0

	if len(sl1) <= len(sl2) {
		for _, val := range sl1 {
			if !StringInSlice(val, sl2) {
				diffCounter++
			}
		}
		diffCounter += len(sl2) - len(sl1)
	} else {
		for _, val := range sl2 {
			if !StringInSlice(val, sl1) {
				diffCounter++
			}
		}
		diffCounter += len(sl1) - len(sl2)
	}

	return diffCounter
}

// GenerateRandomString generates a random string with a lenght of n.
func GenerateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
