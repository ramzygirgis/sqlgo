package codegen

import "testing"

func TestSingularWord(t *testing.T) {
	tests := []struct {
		in, want string
		rule     string // which rule this exercises
	}{
		// plain -s
		{"users", "user", "3f"},
		{"posts", "post", "3f"},
		{"houses", "house", "3f, not the -uses irregulars"},
		{"causes", "cause", "3f, not the -uses irregulars"},
		{"sizes", "size", "3f, not -zzes"},
		{"archives", "archive", "3f, not a -ves rule"},
		{"shoes", "shoe", "3f, not an -oes rule"},
		{"cases", "case", "3f, not the -ases irregulars"},
		{"courses", "course", "3f"},
		{"areas", "area", "3f"},

		// -ies
		{"categories", "category", "3b"},
		{"companies", "company", "3b"},
		{"queries", "query", "3b"},
		{"ties", "tie", "3a"},
		{"pies", "pie", "3a"},

		// -es after sibilants
		{"addresses", "address", "3c sses"},
		{"classes", "class", "3c sses"},
		{"wishes", "wish", "3c shes"},
		{"matches", "match", "3c ches"},
		{"boxes", "box", "3c xes"},
		{"taxes", "tax", "3c xes"},
		{"quizzes", "quiz", "3d"},

		// already singular
		{"status", "status", "3e us"},
		{"address", "address", "3e ss"},
		{"analysis", "analysis", "3e is"},
		{"bus", "bus", "3e us"},

		// not plural at all
		{"person", "person", "3g"},
		{"child", "child", "3g"},
		{"user", "user", "3g"},

		// irregulars
		{"people", "person", "irregular"},
		{"children", "child", "irregular"},
		{"men", "man", "irregular"},
		{"mice", "mouse", "irregular"},
		{"criteria", "criterion", "irregular"},
		{"indices", "index", "irregular"},
		{"matrices", "matrix", "irregular"},
		{"analyses", "analysis", "irregular"},
		{"statuses", "status", "irregular -uses"},
		{"buses", "bus", "irregular -uses"},
		{"aliases", "alias", "irregular -ases"},
		{"knives", "knife", "irregular -ves"},
		{"shelves", "shelf", "irregular -ves"},
		{"heroes", "hero", "irregular -oes"},
		{"caches", "cache", "irregular -ches"},
		{"movies", "movie", "irregular -ies"},
		{"cookies", "cookie", "irregular -ies"},

		// uncountables
		{"data", "data", "uncountable"},
		{"news", "news", "uncountable"},
		{"series", "series", "uncountable"},
		{"species", "species", "uncountable"},
		{"staff", "staff", "uncountable"},

		// edge
		{"", "", "empty"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := singularWord(tt.in); got != tt.want {
				t.Errorf("singularWord(%q) = %q, want %q (rule %s)", tt.in, got, tt.want, tt.rule)
			}
		})
	}
}

func TestSingularize(t *testing.T) {
	tests := []struct{ in, want string }{
		{"users", "user"},
		{"user_accounts", "user_account"},
		{"user_statuses", "user_status"},
		{"news_items", "news_item"}, // only the last word is singularized
		{"account_data", "account_data"},
		{"blog_post_categories", "blog_post_category"},
		{"Users", "user"}, // lowercased
		{"USER_ACCOUNTS", "user_account"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := singularize(tt.in); got != tt.want {
				t.Errorf("singularize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// Every irregular maps to its listed singular, and every uncountable is unchanged.
func TestWordListsAreHonored(t *testing.T) {
	for plural, want := range irregulars {
		if got := singularWord(plural); got != want {
			t.Errorf("irregular %q: got %q, want %q", plural, got, want)
		}
	}
	for word := range uncountables {
		if got := singularWord(word); got != word {
			t.Errorf("uncountable %q: got %q, want unchanged", word, got)
		}
	}
}

// Singularizing a singular word should not change it, so a table that is
// already named in the singular ("user") keeps its name.
func TestSingularWordIsIdempotent(t *testing.T) {
	words := []string{
		"user", "post", "house", "category", "company", "tie", "address",
		"class", "wish", "match", "box", "quiz", "status", "analysis",
		"bus", "person", "child", "criterion", "index", "movie", "cache",
		"knife", "hero", "data", "news", "series",
	}
	for _, w := range words {
		once := singularWord(w)
		if twice := singularWord(once); twice != once {
			t.Errorf("not idempotent: %q -> %q -> %q", w, once, twice)
		}
		if once != w {
			t.Errorf("singular word %q changed to %q", w, once)
		}
	}
}
