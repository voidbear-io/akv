package cmd

import "testing"

func TestReleaseAPIURL(t *testing.T) {
	got := releaseAPIURL("frostyeti", "akv", "latest", false)
	if got != "https://api.github.com/repos/frostyeti/akv/releases/latest" {
		t.Fatalf("unexpected latest url: %s", got)
	}
}
