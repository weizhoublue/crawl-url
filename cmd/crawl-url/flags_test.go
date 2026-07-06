package main

import "testing"

func TestParseCLIRejectsSingleDashLongOption(t *testing.T) {
	_, err := parseCLI([]string{"https://example.com/docs", "-url-limit", "5"})
	if err == nil {
		t.Fatal("expected error for -url-limit")
	}
}

func TestParseCLIRejectsUnknownOption(t *testing.T) {
	_, err := parseCLI([]string{"https://example.com/docs", "--hi"})
	if err == nil {
		t.Fatal("expected error for --hi")
	}
}

func TestParseCLIDoubleDash(t *testing.T) {
	opts, err := parseCLI([]string{"https://example.com/docs", "--url-limit", "5", "--debug"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.cfg.URLLimit == nil || *opts.cfg.URLLimit != 5 {
		t.Fatalf("limit: %#v", opts.cfg.URLLimit)
	}
	if !opts.cfg.Debug {
		t.Fatal("expected debug")
	}
}

func TestParseCLIHelp(t *testing.T) {
	_, err := parseCLI([]string{"https://example.com/docs", "--help"})
	if err != errHelp {
		t.Fatalf("expected errHelp, got %v", err)
	}
}

func TestParseCLITxt(t *testing.T) {
	opts, err := parseCLI([]string{"https://example.com/docs", "--txt"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.cfg.Txt {
		t.Fatal("expected cfg.Txt = true")
	}
}

func TestParseCLITxtRejectsValue(t *testing.T) {
	_, err := parseCLI([]string{"https://example.com/docs", "--txt=foo"})
	if err == nil {
		t.Fatal("expected error for --txt=foo")
	}
}

func TestParseCLIVersion(t *testing.T) {
	_, err := parseCLI([]string{"--version"})
	if err != errVersion {
		t.Fatalf("expected errVersion, got %v", err)
	}
}

func TestParseCLIVersionShort(t *testing.T) {
	_, err := parseCLI([]string{"-v"})
	if err != errVersion {
		t.Fatalf("expected errVersion for -v, got %v", err)
	}
}

func TestParseCLIVersionBeforeSeed(t *testing.T) {
	_, err := parseCLI([]string{"https://example.com/docs", "--version"})
	if err != errVersion {
		t.Fatalf("expected errVersion when --version appears after seed, got %v", err)
	}
}
