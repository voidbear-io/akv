package cmd

import "testing"

func TestReleaseAPIURL(t *testing.T) {
	got := releaseAPIURL("voidbear-io", "akv", "latest", false)
	if got != "https://api.github.com/repos/voidbear-io/akv/releases/latest" {
		t.Fatalf("unexpected latest url: %s", got)
	}
}
