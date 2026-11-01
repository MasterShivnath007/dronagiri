package model

import "testing"

func TestFinishBuildsReverseDependencies(t *testing.T) {
	tree := &Inventory{Tools: []*Tool{
		{ID: "brew:ffmpeg", Name: "ffmpeg", UserInstalled: true,
			Dependencies: []string{"brew:x265", "brew:aom"}},
		{ID: "brew:x265", Name: "x265", Dependencies: []string{"brew:libnuma"}},
		{ID: "brew:aom", Name: "aom"},
		{ID: "brew:libnuma", Name: "libnuma"},
	}}
	tree.Finish()

	cases := map[string][]string{
		"brew:x265":    {"ffmpeg"},
		"brew:aom":     {"ffmpeg"},
		"brew:libnuma": {"ffmpeg"}, // reached transitively
	}
	for id, want := range cases {
		got := tree.RequiredBy[id]
		if len(got) != len(want) || (len(got) > 0 && got[0] != want[0]) {
			t.Errorf("RequiredBy[%s] = %v, want %v", id, got, want)
		}
	}
	if _, ok := tree.RequiredBy["brew:ffmpeg"]; ok {
		t.Error("a package must not be recorded as requiring itself")
	}
}

func TestFinishSortsUserPackagesFirst(t *testing.T) {
	inv := &Inventory{Tools: []*Tool{
		{ID: "brew:zzz", Name: "zzz", UserInstalled: false},
		{ID: "brew:aaa", Name: "aaa", UserInstalled: true},
	}}
	inv.Finish()

	if !inv.Tools[0].UserInstalled {
		t.Errorf("expected user-installed package first, got %s", inv.Tools[0].Name)
	}
}

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{999, "999 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1536 * 1024, "1.5 MB"},
	}
	for _, c := range cases {
		if got := HumanBytes(c.bytes); got != c.want {
			t.Errorf("HumanBytes(%d) = %q, want %q", c.bytes, got, c.want)
		}
	}
}

func TestOutdated(t *testing.T) {
	tool := &Tool{Version: "1.0.0"}
	if tool.Outdated() {
		t.Error("a package with no known latest version is not outdated")
	}
	tool.Latest = "1.1.0"
	if !tool.Outdated() {
		t.Error("expected outdated when latest differs")
	}
	tool.Latest = "1.0.0"
	if tool.Outdated() {
		t.Error("equal versions are not outdated")
	}
}
