package email

import (
	"reflect"
	"testing"
)

func TestParseValid(t *testing.T) {
	cases := []struct {
		in       string
		wantAddr string
		wantName string
	}{
		{"ada@example.com", "ada@example.com", ""},
		{"  ada@example.com  ", "ada@example.com", ""},
		{"Ada Lovelace <ada@example.com>", "ada@example.com", "Ada Lovelace"},
		{"\"Lovelace, Ada\" <ada@example.com>", "ada@example.com", "Lovelace, Ada"},
	}
	for _, c := range cases {
		a, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", c.in, err)
			continue
		}
		if a.Address != c.wantAddr || a.Name != c.wantName {
			t.Errorf("Parse(%q) = %+v, want addr=%q name=%q", c.in, a, c.wantAddr, c.wantName)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{"", "not-an-email", "foo@", "@bar.com", "foo@bar", "a@b@c.com", "plain.text"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) expected error, got nil", in)
		}
	}
}

func TestParseListDelimiters(t *testing.T) {
	valid, invalid := ParseList("a@x.com, b@x.com; c@x.com,not-valid")
	if len(valid) != 3 {
		t.Errorf("expected 3 valid, got %d (%+v)", len(valid), valid)
	}
	if len(invalid) != 1 || invalid[0] != "not-valid" {
		t.Errorf("expected 1 invalid 'not-valid', got %v", invalid)
	}
}

func TestParseListWithDisplayNamesAndCommas(t *testing.T) {
	valid, invalid := ParseList("\"Doe, John\" <john@x.com>; jane@x.com")
	if len(invalid) != 0 {
		t.Errorf("unexpected invalid: %v", invalid)
	}
	if len(valid) != 2 {
		t.Fatalf("expected 2 valid, got %d (%+v)", len(valid), valid)
	}
	if valid[0].Name != "Doe, John" {
		t.Errorf("display name with comma not preserved: %q", valid[0].Name)
	}
}

func TestNormalizeKeyPolicy(t *testing.T) {
	if NormalizeKey("Alice@Example.com") != NormalizeKey("alice@example.com") {
		t.Error("case should be normalized")
	}
	// Plus tags and dots must NOT be normalized away.
	if NormalizeKey("alice+news@example.com") == NormalizeKey("alice@example.com") {
		t.Error("plus tag should not be stripped")
	}
	if NormalizeKey("a.lice@example.com") == NormalizeKey("alice@example.com") {
		t.Error("dots should not be stripped")
	}
}

func TestResolveDuplicatesKeepFirst(t *testing.T) {
	emails := []string{"a@x.com", "b@x.com", "A@x.com", "c@x.com", "b@x.com"}
	res := ResolveDuplicates(emails, PolicyKeepFirst)
	if !reflect.DeepEqual(res.Keep, []int{0, 1, 3}) {
		t.Errorf("Keep = %v, want [0 1 3]", res.Keep)
	}
	if !reflect.DeepEqual(res.Dropped, []int{2, 4}) {
		t.Errorf("Dropped = %v, want [2 4]", res.Dropped)
	}
	if len(res.Groups) != 2 {
		t.Errorf("expected 2 duplicate groups, got %d", len(res.Groups))
	}
}

func TestResolveDuplicatesPolicies(t *testing.T) {
	emails := []string{"a@x.com", "A@x.com", "b@x.com"}
	if got := ResolveDuplicates(emails, PolicyKeepLast).Keep; !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("KeepLast Keep = %v, want [1 2]", got)
	}
	if got := ResolveDuplicates(emails, PolicyExcludeAll).Keep; !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("ExcludeAll Keep = %v, want [2]", got)
	}
	if got := ResolveDuplicates(emails, PolicyKeepAll).Keep; !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Errorf("KeepAll Keep = %v, want [0 1 2]", got)
	}
	man := ResolveDuplicates(emails, PolicyManual)
	if !man.NeedsManualReview {
		t.Error("manual policy with duplicates should need review")
	}
}

func TestResolveDuplicatesInvalidPolicyFallsBack(t *testing.T) {
	res := ResolveDuplicates([]string{"a@x.com", "a@x.com"}, Policy("bogus"))
	if !reflect.DeepEqual(res.Keep, []int{0}) {
		t.Errorf("invalid policy should fall back to keep-first, got %v", res.Keep)
	}
}
