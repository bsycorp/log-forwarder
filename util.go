package main

import "strings"

// All this stuff is O(n) or O(n^2) and we don't care.

// Returns true iff haystack contains needle
func ListContains(haystack []string, needle string) bool {
	for _, straw := range haystack {
		if straw == needle {
			return true
		}
	}
	return false
}

func MapKeysContains(haystack map[string]string, needle string) bool {
	for straw := range haystack {
		if straw == needle {
			return true
		}
	}
	return false
}

// Returns A intersect B
func ListIntersect(a []string, b []string) []string {
	r := []string{}
	for _, s := range a {
		if ListContains(b, s) {
			r = append(r, s)
		}
	}
	return r
}

// Returns A - B
func ListSubtract(a []string, b []string) []string {
	r := []string{}
	for _, s := range a {
		if !ListContains(b, s) {
			r = append(r, s)
		}
	}
	return r
}

// When strings.Split("", sep)  it gets an empty input, it returns [""].
// What I almost always want is for it to return [] in this case.
func Split(s string, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}
