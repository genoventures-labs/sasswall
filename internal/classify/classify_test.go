package classify

import (
	"net/http/httptest"
	"testing"
)

func TestClassifyHoneyBeatsScanner(t *testing.T) {
	c, err := New([]string{"/.env"}, "")
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "http://x/.env", nil)
	r.Header.Set("User-Agent", "nmap")
	res := c.Classify(r, false)
	if res.Category != CategoryHoney {
		t.Fatalf("expected honey, got %s", res.Category)
	}
}

func TestClassifyDenied(t *testing.T) {
	c, _ := New(nil, "")
	r := httptest.NewRequest("GET", "http://x/", nil)
	res := c.Classify(r, true)
	if res.Category != CategoryDenied {
		t.Fatalf("expected denied")
	}
}
