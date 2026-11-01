package source

import (
	"testing"

	"github.com/MasterShivnath007/dronagiri/internal/model"
)

// A gem listed only as `(default: 1.2.3)` ships with Ruby itself and must not
// appear as something the user installed.
func TestPickGemVersionSkipsRubyDefaults(t *testing.T) {
	if _, onlyDefault := pickGemVersion("default: 1.4.1"); !onlyDefault {
		t.Error("expected default-only gem to be skipped")
	}
}

func TestPickGemVersionChoosesNewest(t *testing.T) {
	version, onlyDefault := pickGemVersion("1.2.3, 1.10.0, 1.9.9")
	if onlyDefault {
		t.Fatal("expected a real version")
	}
	if version != "1.10.0" {
		t.Errorf("got %q, want numeric ordering to pick 1.10.0", version)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.10.0", "1.9.9", 1},
		{"1.0.0", "1.0.0", 0},
		{"0.9", "1.0", -1},
		{"2.0.1", "2.0", 1},
	}
	for _, c := range cases {
		if got := model.CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestDedupe(t *testing.T) {
	got := dedupe([]string{"a", "", "b", "a"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("dedupe = %v", got)
	}
}
