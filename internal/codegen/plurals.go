package codegen

import "strings"

// this is AI gen, there is no way i'd write out every plural -> singular conversion LMAO

// uncountables are returned unchanged.
var uncountables = map[string]struct{}{
	"data": {}, "metadata": {}, "news": {}, "series": {}, "species": {},
	"sheep": {}, "fish": {}, "equipment": {}, "information": {},
	"feedback": {}, "staff": {}, "software": {}, "hardware": {}, "media": {},
}

// irregulars maps plurals that no suffix rule handles correctly.
var irregulars = map[string]string{
	// truly irregular
	"people":   "person",
	"men":      "man",
	"women":    "woman",
	"children": "child",
	"teeth":    "tooth",
	"feet":     "foot",
	"mice":     "mouse",
	"geese":    "goose",

	// Greek / Latin
	"criteria":  "criterion",
	"phenomena": "phenomenon",
	"indices":   "index",
	"matrices":  "matrix",
	"vertices":  "vertex",
	"analyses":  "analysis",
	"crises":    "crisis",
	"theses":    "thesis",

	// -uses (a general rule would break "houses", "causes")
	"statuses": "status",
	"campuses": "campus",
	"buses":    "bus",
	"bonuses":  "bonus",
	"viruses":  "virus",
	"censuses": "census",

	// -ases (a general rule would break "cases", "bases")
	"aliases":  "alias",
	"canvases": "canvas",

	// -ves (a general rule would break "archives", "drives")
	"knives":  "knife",
	"wives":   "wife",
	"lives":   "life",
	"shelves": "shelf",
	"halves":  "half",
	"wolves":  "wolf",
	"leaves":  "leaf",
	"thieves": "thief",

	// -oes (a general rule would break "shoes")
	"heroes":   "hero",
	"potatoes": "potato",
	"echoes":   "echo",
	"vetoes":   "veto",

	// -ches where the singular ends in e
	"caches": "cache",
	"niches": "niche",

	// -ies where the singular ends in ie
	"movies":  "movie",
	"cookies": "cookie",
	"zombies": "zombie",
}

// singularize turns a snake_case table name into its singular form by
// singularizing only the last word: "user_accounts" -> "user_account".
// The input is lowercased; casing is the naming step's job.
func singularize(name string) string {
	name = strings.ToLower(name)
	i := strings.LastIndexByte(name, '_')
	prefix, last := name[:i+1], name[i+1:] // i == -1 gives prefix "" and last = name
	return prefix + singularWord(last)
}

// singularWord applies the rules to a single lowercase word.
// Order matters: the first matching rule wins.
func singularWord(w string) string {
	if w == "" {
		return w
	}

	// 1. uncountables
	if _, ok := uncountables[w]; ok {
		return w
	}

	// 2. irregulars
	if s, ok := irregulars[w]; ok {
		return s
	}

	// 3. suffix rules
	switch {
	// 3a. short -ies words: ties -> tie, pies -> pie
	case strings.HasSuffix(w, "ies") && len(w) <= 4:
		return strings.TrimSuffix(w, "s")

	// 3b. -ies -> -y: categories -> category
	case strings.HasSuffix(w, "ies"):
		return strings.TrimSuffix(w, "ies") + "y"

	// 3c. -sses, -shes, -ches, -xes -> drop "es"
	case strings.HasSuffix(w, "sses"),
		strings.HasSuffix(w, "shes"),
		strings.HasSuffix(w, "ches"),
		strings.HasSuffix(w, "xes"):
		return strings.TrimSuffix(w, "es")

	// 3d. -zzes -> drop "zes": quizzes -> quiz
	case strings.HasSuffix(w, "zzes"):
		return strings.TrimSuffix(w, "zes")

	// 3e. already singular: address, status, analysis
	case strings.HasSuffix(w, "ss"),
		strings.HasSuffix(w, "us"),
		strings.HasSuffix(w, "is"):
		return w

	// 3f. plain -s: users -> user, houses -> house
	case strings.HasSuffix(w, "s"):
		return strings.TrimSuffix(w, "s")

	// 3g. not plural
	default:
		return w
	}
}
