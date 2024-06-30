package git

import (
	"regexp"
	"strings"
)

func globMatch(pattern, s string) bool {
	escaped := regexp.QuoteMeta(pattern)
	// Replace escaped * with .* to make it a regexp pattern.
	pattern = strings.ReplaceAll(escaped, "\\*", ".*")
	matcher, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return matcher.MatchString(s)
}
