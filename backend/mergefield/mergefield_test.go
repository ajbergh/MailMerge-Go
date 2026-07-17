package mergefield

import (
	"testing"

	"MailMergeApp/backend/models"
)

func TestCanonicalize(t *testing.T) {
	cases := map[string]string{
		"First Name":      "first_name",
		"Account-Manager": "account_manager",
		"Customer ID":     "customer_id",
		"E-mail Address":  "e_mail_address",
		"  Spaced  ":      "spaced",
		"already_snake":   "already_snake",
		"FirstName":       "firstname",
		"Präsident":       "präsident",
		"a---b":           "a_b",
		"!!!":             "",
	}
	for in, want := range cases {
		if got := Canonicalize(in); got != want {
			t.Errorf("Canonicalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNewSchemaStandardFields(t *testing.T) {
	s := NewSchema(nil)
	for _, id := range []string{StandardFirstName, StandardLastName, StandardEmail} {
		if _, ok := s.Resolve(id); !ok {
			t.Errorf("standard field %q not resolvable", id)
		}
	}
	// Legacy single-word aliases must resolve to standard fields.
	for _, alias := range []string{"FirstName", "First Name", "firstname", "LastName", "Email", "email address"} {
		if _, ok := s.Resolve(alias); !ok {
			t.Errorf("alias %q did not resolve to a standard field", alias)
		}
	}
}

func TestNewSchemaCustomHeaders(t *testing.T) {
	s := NewSchema([]string{"First Name", "Account-Manager", "Customer ID", "E-mail Address", "Región"})

	f, ok := s.Resolve("Account-Manager")
	if !ok {
		t.Fatal("Account-Manager not resolvable")
	}
	if f.ID != "account_manager" {
		t.Errorf("got ID %q, want account_manager", f.ID)
	}
	if f.Token() != "{{account_manager}}" {
		t.Errorf("got token %q", f.Token())
	}

	// "E-mail Address" canonicalizes to e_mail_address which is a standard email alias,
	// so it should fold into the standard email field, not create a new one.
	if f, ok := s.Resolve("E-mail Address"); !ok || f.ID != StandardEmail {
		t.Errorf("E-mail Address should fold into standard email field, got %v ok=%v", f.ID, ok)
	}

	// Unicode header
	if f, ok := s.Resolve("Región"); !ok || f.ID != "región" {
		t.Errorf("unicode header not handled: %v ok=%v", f.ID, ok)
	}
}

func TestCollisionDetection(t *testing.T) {
	// "Customer ID" and "customer-id" both canonicalize to customer_id.
	s := NewSchema([]string{"Customer ID", "customer-id", "Company"})
	if len(s.Collisions) != 1 {
		t.Fatalf("expected 1 collision, got %d: %+v", len(s.Collisions), s.Collisions)
	}
	if s.Collisions[0].ID != "customer_id" {
		t.Errorf("collision ID = %q, want customer_id", s.Collisions[0].ID)
	}
	if len(s.Collisions[0].DisplayNames) != 2 {
		t.Errorf("expected 2 colliding display names, got %v", s.Collisions[0].DisplayNames)
	}
	if !s.HasCollision("customer_id") {
		t.Error("HasCollision(customer_id) = false")
	}
	if s.HasCollision("company") {
		t.Error("Company should not be a collision")
	}
}

func TestValueFor(t *testing.T) {
	c := models.Contact{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
		CustomFields: map[string]string{
			"account-manager": "Bob",
			"customer id":     "",
		},
	}
	s := NewSchema([]string{"Account-Manager", "Customer ID"})

	fn, _ := s.Resolve("first_name")
	if v, ok := ValueFor(fn, c); v != "Ada" || !ok {
		t.Errorf("first_name = %q ok=%v", v, ok)
	}

	am, _ := s.Resolve("account_manager")
	if v, ok := ValueFor(am, c); v != "Bob" || !ok {
		t.Errorf("account_manager = %q ok=%v", v, ok)
	}

	cid, _ := s.Resolve("customer_id")
	if v, ok := ValueFor(cid, c); v != "" || ok {
		t.Errorf("customer_id present-but-blank = %q ok=%v, want empty and false", v, ok)
	}
}
