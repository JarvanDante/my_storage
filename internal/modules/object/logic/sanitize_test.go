package logic

import "testing"

func TestSanitizeFilenameKeepsChineseAndBnc(t *testing.T) {
	got := sanitizeFilename("照片.jpg")
	if got != "照片.jpg" {
		t.Fatalf("got %q", got)
	}
	got = sanitizeFilename("微信图片_2026.bnc")
	if got != "微信图片_2026.bnc" {
		t.Fatalf("got %q", got)
	}
	got = sanitizeFilename(".bnc")
	if got != "file.bnc" {
		t.Fatalf("got %q", got)
	}
}
