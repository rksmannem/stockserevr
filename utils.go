package main

import "strings"

// Find ...
func Find(a []string, x string) bool {
	for _, n := range a {
		if strings.TrimSpace(x) == strings.TrimSpace(n) {
			return true
		}
	}
	return false
}
