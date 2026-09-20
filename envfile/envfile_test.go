package envfile

import (
	"maps"
	"os"
	"strings"
	"testing"
)

func parse(t *testing.T, src string) (map[string]string, error) {
	t.Helper()

	out := map[string]string{}

	return out, parseContent([]byte(src), out)
}

func TestParseStatements(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want map[string]string
	}{
		{"basic", "FOO=bar\nBAZ=qux\n", map[string]string{"FOO": "bar", "BAZ": "qux"}},
		{"colon separator", "FOO: bar\nBAZ: qux\n", map[string]string{"FOO": "bar", "BAZ": "qux"}},
		{"no trailing newline", "FOO=bar\nBAZ=qux", map[string]string{"FOO": "bar", "BAZ": "qux"}},
		{"crlf", "FOO=bar\r\nBAZ=qux\r\n", map[string]string{"FOO": "bar", "BAZ": "qux"}},
		{"lone cr", "FOO=bar\rBAZ=qux\r", map[string]string{"FOO": "bar", "BAZ": "qux"}},
		{"nothing to parse", "\ufeff# one\n  # two\n\n   \t\n", map[string]string{}},
		{"comment without newline", "# trailing comment", map[string]string{}},
		{"multiple equals", "KEY=foo=bar\n", map[string]string{"KEY": "foo=bar"}},
		{"empty values", "FOO=\nBAR=   \nBAZ=\"\"\nQUX=''\n", map[string]string{"FOO": "", "BAR": "", "BAZ": "", "QUX": ""}},
		{"spaces around key", "  FOO = bar  \n", map[string]string{"FOO": "bar"}},
		{"space inside key", "MY KEY=value\n", map[string]string{"MY KEY": "value"}},
		{"tab inside key", "MY\tKEY=value\n", map[string]string{"MY\tKEY": "value"}},
		{"dots and digits in key", "MY.KEY=1\nKEY123=2\n1KEY=3\n", map[string]string{"MY.KEY": "1", "KEY123": "2", "1KEY": "3"}},
		{
			"unicode key and value", "\u00dcSER=h\u00e9llo\n\u041f\u0410\u0420\u041e\u041b\u042c=\u0441\u0435\u043a\u0440\u0435\u0442\n",
			map[string]string{"\u00dcSER": "h\u00e9llo", "\u041f\u0410\u0420\u041e\u041b\u042c": "\u0441\u0435\u043a\u0440\u0435\u0442"},
		},
		{"utf8 bom", "\ufeffFOO=bar\n", map[string]string{"FOO": "bar"}},
		{
			"export prefix", "export FOO=bar\nexport   BAZ=qux\nexport\tQUX=1\n",
			map[string]string{"FOO": "bar", "BAZ": "qux", "QUX": "1"},
		},
		{"export as key", "export=value\n", map[string]string{"export": "value"}},
		{"inline comment", "FOO=value # comment\n", map[string]string{"FOO": "value"}},
		{"inline comment with tab", "FOO=value\t# comment\n", map[string]string{"FOO": "value"}},
		{"hash without leading space is kept", "FOO=value#notacomment\n", map[string]string{"FOO": "value#notacomment"}},
		{"hash at value start is kept", "FOO=#not-comment\n", map[string]string{"FOO": "#not-comment"}},
		{"comment with hash inside", "FOO=hash # not # real\n", map[string]string{"FOO": "hash"}},
		{"hex color is cut", "FOO=red #ff0000\n", map[string]string{"FOO": "red"}},
		{"quoted value with trailing comment", "FOO=\"val\" # comment\n", map[string]string{"FOO": "val"}},
		{"quotes inside unquoted value", "FOO=he said \"hi\"\n", map[string]string{"FOO": `he said "hi"`}},
		{
			"colons and exclamation in values", "DSN=postgres://user:pass@db:5432/mydb?sslmode=disable\nROOM=!random:example.com\n",
			map[string]string{"DSN": "postgres://user:pass@db:5432/mydb?sslmode=disable", "ROOM": "!random:example.com"},
		},
		{
			"realistic values", "ALLOWLIST=*@example.com *.test.com user@example.net\nENDPOINT=https://key@host.io/123\n" +
				"DISPLAY=**{{ .name }} by {{ .email }}**\nTHROTTLE=5r/m\nIPS=10.0.0.1 192.168.1.0/24 ::1\n",
			map[string]string{
				"ALLOWLIST": "*@example.com *.test.com user@example.net", "ENDPOINT": "https://key@host.io/123",
				"DISPLAY": "**{{ .name }} by {{ .email }}**", "THROTTLE": "5r/m", "IPS": "10.0.0.1 192.168.1.0/24 ::1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := parse(t, tt.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !maps.Equal(out, tt.want) {
				t.Errorf("got %v, want %v", out, tt.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr string
		want    map[string]string
	}{
		{
			"invalid key character", "A=1\nBAD%KEY=x\nB=2\n", "line 2: invalid character '%' in variable name",
			map[string]string{"A": "1", "B": "2"},
		},
		{"hyphen in key", "MY-KEY=1\nB=2\n", "line 1: invalid character '-' in variable name", map[string]string{"B": "2"}},
		{
			"unterminated double quote", "A=1\nQ=\"oops\nB=2\n", "line 2: unterminated double-quoted value",
			map[string]string{"A": "1", "B": "2"},
		},
		{"unterminated single quote", "Q='oops\nB=2\n", "line 1: unterminated single-quoted value", map[string]string{"B": "2"}},
		{"multiline double quote", "Q=\"multi\nline\"\nB=2\n", "line 1: unterminated double-quoted value", map[string]string{"B": "2"}},
		{
			"missing separator mid file", "A=1\nPOSTGRES_PASSWORD\nB=2\n", "line 2: missing '=' or ':' separator",
			map[string]string{"A": "1", "B": "2"},
		},
		{
			"missing separator at eof", "A=1\nPOSTGRES_PASSWORD", "line 2: missing '=' or ':' separator",
			map[string]string{"A": "1"},
		},
		{"export without separator", "export\nB=2\n", "line 1: missing '=' or ':' separator", map[string]string{"B": "2"}},
		{"empty name", "=value\nB=2\n", "line 1: empty variable name", map[string]string{"B": "2"}},
		{
			"trailing content after quote", "A=\"x\" junk\nB=2\n", "line 1: unexpected 'j' after quoted value",
			map[string]string{"B": "2"},
		},
		{
			"escaped quote in single quotes", "A='it\\'s'\nB=2\n", "line 1: unexpected 's' after quoted value",
			map[string]string{"B": "2"},
		},
		{"nul byte in value", "A=with\x00nul\nB=2\n", "line 1: value of A contains a NUL byte", map[string]string{"B": "2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := parse(t, tt.src)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
			if !maps.Equal(out, tt.want) {
				t.Errorf("got %v, want %v", out, tt.want)
			}
		})
	}
}

func TestParseErrorsStayBounded(t *testing.T) {
	src := "A=\"unterminated\nSECRET=super-secret-value\nBAZ=another-secret\n"
	_, err := parse(t, src)
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "super-secret-value") || strings.Contains(err.Error(), "another-secret") {
		t.Errorf("error leaks neighboring values: %q", err)
	}
}

func TestParseMultipleErrors(t *testing.T) {
	out, err := parse(t, "BAD%KEY=x\nA=1\nMY-KEY=2\n")
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"line 1: invalid character '%'", "line 3: invalid character '-'"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err, want)
		}
	}
	if out["A"] != "1" {
		t.Errorf("A = %q, want %q", out["A"], "1")
	}
}

func TestLoadSkipsMissingFilesAndDirectories(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := Load(); err != nil {
		t.Errorf("missing .env must be silent, got %v", err)
	}
	if err := Load("absent.env"); err != nil {
		t.Errorf("missing additional file must be silent, got %v", err)
	}

	if err := os.Mkdir("dir.env", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Load("dir.env"); err != nil {
		t.Errorf("directory must be skipped, got %v", err)
	}
}

func TestLoadSetsVars(t *testing.T) {
	t.Chdir(t.TempDir())
	writeEnvFile(t, DefaultFile, "FOO=bar\nBAZ=qux\n")
	unsetEnv(t, "FOO", "BAZ")

	if err := Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkEnv(t, "FOO", "bar")
	checkEnv(t, "BAZ", "qux")
}

func TestLoadLaterFilesWin(t *testing.T) {
	t.Chdir(t.TempDir())
	writeEnvFile(t, DefaultFile, "FOO=first\n")
	writeEnvFile(t, "a.env", "FOO=second\n")
	writeEnvFile(t, "b.env", "FOO=third\n")
	unsetEnv(t, "FOO")

	if err := Load("a.env", "b.env"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkEnv(t, "FOO", "third")
}

func TestLoadOverridesProcessEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("INHERITED", "from-process")
	writeEnvFile(t, DefaultFile, "INHERITED=from-file\n")

	if err := Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkEnv(t, "INHERITED", "from-file")
}

func TestLoadNeverReadsProcessEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("REAL_SECRET", "s3cr3t")
	writeEnvFile(t, DefaultFile, "DERIVED=$REAL_SECRET\n")
	unsetEnv(t, "DERIVED")

	if err := Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checkEnv(t, "DERIVED", "$REAL_SECRET")
}

func TestLoadKeepsGoodLinesOnError(t *testing.T) {
	t.Chdir(t.TempDir())
	writeEnvFile(t, DefaultFile, "GOOD_ONE=1\nBAD%KEY=x\nGOOD_TWO=2\n")
	unsetEnv(t, "GOOD_ONE", "GOOD_TWO")

	err := Load()
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), ".env: line 2: invalid character '%'") {
		t.Errorf("error = %q, want file name and line number", err)
	}

	checkEnv(t, "GOOD_ONE", "1")
	checkEnv(t, "GOOD_TWO", "2")
	if _, ok := os.LookupEnv("BAD%KEY"); ok {
		t.Error("invalid key must not be set")
	}
}

func TestLoadReportsEveryBrokenFile(t *testing.T) {
	t.Chdir(t.TempDir())
	writeEnvFile(t, DefaultFile, "A=1\nBAD%KEY=x\n")
	writeEnvFile(t, "extra.env", "=value\n")
	unsetEnv(t, "A")

	err := Load("extra.env")
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{".env: line 2:", "extra.env: line 1:"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to mention %q", err, want)
		}
	}

	checkEnv(t, "A", "1")
}

func writeEnvFile(t *testing.T, name, content string) {
	t.Helper()

	if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()

	for _, key := range keys {
		t.Cleanup(func() { os.Unsetenv(key) })
		os.Unsetenv(key)
	}
}

func checkEnv(t *testing.T, key, want string) {
	t.Helper()

	if got := os.Getenv(key); got != want {
		t.Errorf("%s = %q, want %q", key, got, want)
	}
}
