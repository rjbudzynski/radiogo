package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseArgs_Help(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	showHelp, err := parseArgs([]string{"--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if !showHelp {
		t.Fatal("showHelp = false, want true")
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if out := stdout.String(); !strings.Contains(out, "Usage:") || !strings.Contains(out, "--help") {
		t.Fatalf("stdout = %q, want usage text mentioning --help", out)
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	showHelp, err := parseArgs([]string{"--unknown"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("parseArgs error = nil, want error")
	}
	if showHelp {
		t.Fatal("showHelp = true, want false")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if errText := stderr.String(); !strings.Contains(errText, "flag provided but not defined") || !strings.Contains(errText, "Usage:") {
		t.Fatalf("stderr = %q, want parse error and usage text", errText)
	}
}
