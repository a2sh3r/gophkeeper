package version

import (
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	info := Info()
	if !strings.Contains(info, "Version:") {
		t.Error("Info should contain Version field")
	}

	if !strings.Contains(info, "Build Date:") {
		t.Error("Info should contain Build Date field")
	}

	if !strings.Contains(info, "Git Commit:") {
		t.Error("Info should contain Git Commit field")
	}

	if !strings.Contains(info, "Go Version:") {
		t.Error("Info should contain Go Version field")
	}
}

func TestShortInfo(t *testing.T) {
	shortInfo := ShortInfo()

	if !strings.Contains(shortInfo, "GophKeeper") {
		t.Error("ShortInfo should contain 'GophKeeper'")
	}

	if !strings.Contains(shortInfo, Version) {
		t.Error("ShortInfo should contain version")
	}
}
