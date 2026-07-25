package slugify

import (
	"regexp"
	"strings"
)

var accentReplacer = strings.NewReplacer(
	"à", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"î", "i", "ï", "i",
	"ô", "o", "ö", "o",
	"ù", "u", "û", "u", "ü", "u",
	"ç", "c",
	"œ", "oe", "æ", "ae",
	"ÿ", "y", "ñ", "n",
	"À", "a", "Â", "a", "Ä", "a",
	"É", "e", "È", "e", "Ê", "e", "Ë", "e",
	"Î", "i", "Ï", "i",
	"Ô", "o", "Ö", "o",
	"Ù", "u", "Û", "u", "Ü", "u",
	"Ç", "c",
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
var trimHyphen = regexp.MustCompile(`^-+|-+$`)

func Slugify(input string) string {
	s := accentReplacer.Replace(input)
	s = strings.ToLower(s)
	s = nonAlnum.ReplaceAllString(s, "-")
	s = trimHyphen.ReplaceAllString(s, "")
	if s == "" {
		s = "entree"
	}
	return s
}
