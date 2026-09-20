package envfile

import (
	"os"
	"testing"
)

func checkEnv(t *testing.T, key, want string) {
	t.Helper()
	got := os.Getenv(key)
	if got != want {
		t.Errorf("%s = %q, want %q", key, got, want)
	}
}

func TestParseContent(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte{}, out)
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 0 {
			t.Errorf("expected empty map, got %v", out)
		}
	})

	t.Run("comments only", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("# this is a comment\n# another one\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 0 {
			t.Errorf("expected empty map, got %v", out)
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("   \n\t\n  \n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 0 {
			t.Errorf("expected empty map, got %v", out)
		}
	})

	t.Run("mixed comments and whitespace", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("\n  # comment\n  \n  # another\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 0 {
			t.Errorf("expected empty map, got %v", out)
		}
	})

	t.Run("basic key=value", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=bar\nBAZ=qux\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "bar" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "bar")
		}
		if out["BAZ"] != "qux" {
			t.Errorf("BAZ = %q, want %q", out["BAZ"], "qux")
		}
	})

	t.Run("key:value colon separator", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO:bar\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "bar" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "bar")
		}
	})

	t.Run("no trailing newline", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=bar\nBAZ=qux"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "bar" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "bar")
		}
		if out["BAZ"] != "qux" {
			t.Errorf("BAZ = %q, want %q", out["BAZ"], "qux")
		}
	})

	t.Run("crlf normalization", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=bar\r\nBAZ=qux\r\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "bar" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "bar")
		}
		if out["BAZ"] != "qux" {
			t.Errorf("BAZ = %q, want %q", out["BAZ"], "qux")
		}
	})

	t.Run("multiple = signs", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("KEY=foo=bar\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["KEY"] != "foo=bar" {
			t.Errorf("KEY = %q, want %q", out["KEY"], "foo=bar")
		}
	})

	t.Run("whitespace-only value", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=\nBAR=   \n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "" {
			t.Errorf("FOO = %q, want empty", out["FOO"])
		}
		if out["BAR"] != "" {
			t.Errorf("BAR = %q, want empty", out["BAR"])
		}
	})

	t.Run("space in key name", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("MY KEY=value\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["MY KEY"] != "value" {
			t.Errorf("MY KEY = %q, want %q", out["MY KEY"], "value")
		}
	})

	t.Run("key with dots", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("MY.KEY=value\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["MY.KEY"] != "value" {
			t.Errorf("MY.KEY = %q, want %q", out["MY.KEY"], "value")
		}
	})

	t.Run("key with numbers", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("KEY123=value\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["KEY123"] != "value" {
			t.Errorf("KEY123 = %q, want %q", out["KEY123"], "value")
		}
	})

	t.Run("inline comment", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=value # this is a comment\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "value" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "value")
		}
	})

	t.Run("inline comment no space before hash", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=value#notacomment\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "value#notacomment" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "value#notacomment")
		}
	})

	t.Run("inline comment with tab before hash", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=value\t# comment\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "value" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "value")
		}
	})

	t.Run("export prefix", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("export FOO=bar\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "bar" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "bar")
		}
	})

	t.Run("export as key", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("export=value\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["export"] != "value" {
			t.Errorf("export = %q, want %q", out["export"], "value")
		}
	})

	t.Run("export with extra spaces", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("export   FOO=bar\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "bar" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "bar")
		}
	})

	t.Run("single-quoted value", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO='bar'\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "bar" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "bar")
		}
	})

	t.Run("single-quoted with escaped quote", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO='bar\\'s'\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "bar\\'s" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "bar\\'s")
		}
	})

	t.Run("single-quoted no expansion", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO='$HOME'\nBAR='$(cmd)'\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "$HOME" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "$HOME")
		}
		if out["BAR"] != "$(cmd)" {
			t.Errorf("BAR = %q, want %q", out["BAR"], "$(cmd)")
		}
	})

	t.Run("double-quoted value", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=\"bar\"\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "bar" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "bar")
		}
	})

	t.Run("double-quoted escape sequences", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=\"\\\"bar\\\"\"\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != `"bar"` {
			t.Errorf("FOO = %q, want %q", out["FOO"], `"bar"`)
		}
	})

	t.Run("double-quoted newline escape", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=\"line1\\nline2\"\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "line1\nline2" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "line1\nline2")
		}
	})

	t.Run("double-quoted carriage return escape", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=\"line1\\rline2\"\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "line1\rline2" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "line1\rline2")
		}
	})

	t.Run("double-quoted backslash escape", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=\"a\\\\b\"\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "a\\b" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "a\\b")
		}
	})

	t.Run("variable expansion $VAR", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("A=hello\nB=$A/world\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["A"] != "hello" {
			t.Errorf("A = %q, want %q", out["A"], "hello")
		}
		if out["B"] != "hello/world" {
			t.Errorf("B = %q, want %q", out["B"], "hello/world")
		}
	})

	t.Run("variable expansion ${VAR}", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("A=hello\nB=${A}/world\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["B"] != "hello/world" {
			t.Errorf("B = %q, want %q", out["B"], "hello/world")
		}
	})

	t.Run("self-reference empty", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("A=$A\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["A"] != "" {
			t.Errorf("A = %q, want empty", out["A"])
		}
	})

	t.Run("later variable unavailable", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("A=$B\nB=hello\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["A"] != "" {
			t.Errorf("A = %q, want empty", out["A"])
		}
		if out["B"] != "hello" {
			t.Errorf("B = %q, want %q", out["B"], "hello")
		}
	})

	t.Run("escaped dollar", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=\\$HOME\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "$HOME" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "$HOME")
		}
	})

	t.Run("dollar paren literal", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=$(whoami)\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["FOO"] != "(whoami)" {
			t.Errorf("FOO = %q, want %q", out["FOO"], "(whoami)")
		}
	})

	t.Run("lowercase var no expansion", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("A=hello\nB=$a\n"), out)
		if err != nil {
			t.Fatal(err)
		}
		if out["B"] != "$a" {
			t.Errorf("B = %q, want %q", out["B"], "$a")
		}
	})

	t.Run("unterminated single quote error", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO='bar\n"), out)
		if err == nil {
			t.Fatal("expected error for unterminated quote")
		}
	})

	t.Run("unterminated double quote error", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO=\"bar\n"), out)
		if err == nil {
			t.Fatal("expected error for unterminated quote")
		}
	})

	t.Run("invalid key character", func(t *testing.T) {
		out := map[string]string{}
		err := parseContent([]byte("FOO%BAR=value\n"), out)
		if err == nil {
			t.Fatal("expected error for invalid key char")
		}
	})

	t.Run("golden EXAMPLE_ENV content", func(t *testing.T) {
		content := []byte("TEST_USER=admin\n" +
			"TEST_SECRET=s3kr1t\n" +
			"TEST_HOST=http://localhost:8080\n" +
			"TEST_DSN=postgres://user:pass@db:5432/mydb?sslmode=disable\n" +
			"TEST_DRIVER=postgres\n" +
			"TEST_ALLOWLIST=*@example.com *.test.com user@example.net\n" +
			"TEST_ENDPOINT=https://key@host.io/123\n" +
			"TEST_EMPTY=\n" +
			"TEST_EMPTY2=\n" +
			"TEST_LEVEL=DEBUG\n" +
			"TEST_LIMIT=5000\n" +
			"TEST_IGNORE=\n" +
			"TEST_TOKEN=abc123-def456-ghi789\n" +
			"TEST_SENDER=bot@example.com\n" +
			"TEST_REPLYTO=reply@example.net\n" +
			"TEST_RECIPIENT=alert@example.org\n" +
			"TEST_FLAG=1\n" +
			"TEST_API_USER=apiuserhere\n" +
			"TEST_API_SECRET=apisecrethere\n" +
			"TEST_ALLOWED_IPS=10.0.0.1 192.168.1.0/24 ::1\n" +
			"TEST_SVC_HOST=\n" +
			"TEST_SVC_KEY=\n" +
			"TEST_SVC_NAME=\n" +
			"TEST_SVC_RETRIES=\n" +
			"TEST_SVC_TIMEOUT=\n" +
			"TEST_ROOM_ID=!random:example.com\n" +
			"TEST_DISPLAY=\n" +
			"TEST_REDIRECT=https://example.net\n" +
			"TEST_REDIRECT_BACKUP=\n" +
			"TEST_FEATURE_X=1\n" +
			"TEST_FEATURE_Y=1\n" +
			"TEST_THROTTLE=5r/m\n" +
			"TEST_THROTTLE_SHARED=False\n" +
			"TEST_EXTRA_FIELDS=name\n" +
			"TEST_NOTIFY_SUBJECT=\n" +
			"TEST_NOTIFY_BODY=\n" +
			"TEST_ITEMS=main\n" +
			"\n" +
			"TEST_EMPTY=https://real.url\n" +
			"TEST_EMPTY2=uuid-here\n" +
			"TEST_SVC_HOST=https://svc.example.com\n" +
			"TEST_SVC_KEY=real-api-key-here\n" +
			"TEST_SVC_NAME=production\n" +
			"TEST_SVC_RETRIES=3\n" +
			"TEST_SVC_TIMEOUT=30\n" +
			"TEST_REDIRECT_BACKUP=https://backup.example.de\n" +
			"TEST_DISPLAY=**{{ .name }} by {{ .email }}**\n" +
			"TEST_MASTER_SECRET=master-secret\n" +
			"TEST_SLAVE_SECRET=slave-secret\n")

		out := map[string]string{}
		err := parseContent(content, out)
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			key, want string
		}{
			{"TEST_USER", "admin"},
			{"TEST_SECRET", "s3kr1t"},
			{"TEST_HOST", "http://localhost:8080"},
			{"TEST_DSN", "postgres://user:pass@db:5432/mydb?sslmode=disable"},
			{"TEST_DRIVER", "postgres"},
			{"TEST_ALLOWLIST", "*@example.com *.test.com user@example.net"},
			{"TEST_ENDPOINT", "https://key@host.io/123"},
			{"TEST_EMPTY", "https://real.url"},
			{"TEST_EMPTY2", "uuid-here"},
			{"TEST_LEVEL", "DEBUG"},
			{"TEST_LIMIT", "5000"},
			{"TEST_IGNORE", ""},
			{"TEST_TOKEN", "abc123-def456-ghi789"},
			{"TEST_SENDER", "bot@example.com"},
			{"TEST_REPLYTO", "reply@example.net"},
			{"TEST_RECIPIENT", "alert@example.org"},
			{"TEST_FLAG", "1"},
			{"TEST_API_USER", "apiuserhere"},
			{"TEST_API_SECRET", "apisecrethere"},
			{"TEST_ALLOWED_IPS", "10.0.0.1 192.168.1.0/24 ::1"},
			{"TEST_SVC_HOST", "https://svc.example.com"},
			{"TEST_SVC_KEY", "real-api-key-here"},
			{"TEST_SVC_NAME", "production"},
			{"TEST_SVC_RETRIES", "3"},
			{"TEST_SVC_TIMEOUT", "30"},
			{"TEST_ROOM_ID", "!random:example.com"},
			{"TEST_DISPLAY", "**{{ .name }} by {{ .email }}**"},
			{"TEST_REDIRECT", "https://example.net"},
			{"TEST_REDIRECT_BACKUP", "https://backup.example.de"},
			{"TEST_FEATURE_X", "1"},
			{"TEST_FEATURE_Y", "1"},
			{"TEST_THROTTLE", "5r/m"},
			{"TEST_THROTTLE_SHARED", "False"},
			{"TEST_EXTRA_FIELDS", "name"},
			{"TEST_NOTIFY_SUBJECT", ""},
			{"TEST_NOTIFY_BODY", ""},
			{"TEST_ITEMS", "main"},
			{"TEST_MASTER_SECRET", "master-secret"},
			{"TEST_SLAVE_SECRET", "slave-secret"},
		}
		for _, tt := range tests {
			if out[tt.key] != tt.want {
				t.Errorf("%s = %q, want %q", tt.key, out[tt.key], tt.want)
			}
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("missing .env is silent", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		Load()
	})

	t.Run("additional file missing is silent", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		Load("nonexistent.env")
	})

	t.Run("loads .env and sets env vars", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		os.WriteFile(".env", []byte("FOO=bar\nBAZ=qux\n"), 0o644)
		os.Unsetenv("FOO")
		os.Unsetenv("BAZ")
		Load()
		checkEnv(t, "FOO", "bar")
		checkEnv(t, "BAZ", "qux")
	})

	t.Run("additional files override .env", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		os.WriteFile(".env", []byte("FOO=bar\n"), 0o644)
		os.WriteFile("override.env", []byte("FOO=override\nBAZ=qux\n"), 0o644)
		os.Unsetenv("FOO")
		os.Unsetenv("BAZ")
		Load("override.env")
		checkEnv(t, "FOO", "override")
		checkEnv(t, "BAZ", "qux")
	})

	t.Run("later files overwrite earlier keys", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		os.WriteFile(".env", []byte("FOO=first\n"), 0o644)
		os.WriteFile("a.env", []byte("FOO=second\n"), 0o644)
		os.WriteFile("b.env", []byte("FOO=third\n"), 0o644)
		os.Unsetenv("FOO")
		Load("a.env", "b.env")
		checkEnv(t, "FOO", "third")
	})
}

func TestSkipToStatement(t *testing.T) {
	t.Run("nil on empty", func(t *testing.T) {
		if got := skipToStatement([]byte{}); got != nil {
			t.Errorf("expected nil, got %q", got)
		}
	})

	t.Run("nil on whitespace", func(t *testing.T) {
		if got := skipToStatement([]byte("   \n\t\n  ")); got != nil {
			t.Errorf("expected nil, got %q", got)
		}
	})

	t.Run("skips leading whitespace", func(t *testing.T) {
		got := skipToStatement([]byte("  \n  FOO=bar"))
		if string(got) != "FOO=bar" {
			t.Errorf("expected %q, got %q", "FOO=bar", string(got))
		}
	})

	t.Run("skips comment line", func(t *testing.T) {
		got := skipToStatement([]byte("# comment\nFOO=bar"))
		if string(got) != "FOO=bar" {
			t.Errorf("expected %q, got %q", "FOO=bar", string(got))
		}
	})

	t.Run("nil on eof inside comment", func(t *testing.T) {
		got := skipToStatement([]byte("# comment with no newline"))
		if got != nil {
			t.Errorf("expected nil, got %q", got)
		}
	})
}

func TestExtractKey(t *testing.T) {
	t.Run("basic key=value", func(t *testing.T) {
		key, remaining, err := extractKey([]byte("FOO=bar"))
		if err != nil {
			t.Fatal(err)
		}
		if key != "FOO" {
			t.Errorf("key = %q, want %q", key, "FOO")
		}
		if string(remaining) != "bar" {
			t.Errorf("remaining = %q, want %q", string(remaining), "bar")
		}
	})

	t.Run("space in key", func(t *testing.T) {
		key, remaining, err := extractKey([]byte("MY KEY=value"))
		if err != nil {
			t.Fatal(err)
		}
		if key != "MY KEY" {
			t.Errorf("key = %q, want %q", key, "MY KEY")
		}
		if string(remaining) != "value" {
			t.Errorf("remaining = %q, want %q", string(remaining), "value")
		}
	})

	t.Run("trailing space in key trimmed", func(t *testing.T) {
		key, remaining, err := extractKey([]byte("MY KEY = value"))
		if err != nil {
			t.Fatal(err)
		}
		if key != "MY KEY" {
			t.Errorf("key = %q, want %q", key, "MY KEY")
		}
		if string(remaining) != "value" {
			t.Errorf("remaining = %q, want %q", string(remaining), "value")
		}
	})

	t.Run("export stripped", func(t *testing.T) {
		key, remaining, err := extractKey([]byte("export FOO=bar"))
		if err != nil {
			t.Fatal(err)
		}
		if key != "FOO" {
			t.Errorf("key = %q, want %q", key, "FOO")
		}
		if string(remaining) != "bar" {
			t.Errorf("remaining = %q, want %q", string(remaining), "bar")
		}
	})

	t.Run("export without space is key", func(t *testing.T) {
		key, remaining, err := extractKey([]byte("export=value"))
		if err != nil {
			t.Fatal(err)
		}
		if key != "export" {
			t.Errorf("key = %q, want %q", key, "export")
		}
		if string(remaining) != "value" {
			t.Errorf("remaining = %q, want %q", string(remaining), "value")
		}
	})

	t.Run("error on invalid char", func(t *testing.T) {
		_, _, err := extractKey([]byte("FOO%BAR=value"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestExtractValue(t *testing.T) {
	t.Run("empty src", func(t *testing.T) {
		val, rest, err := extractValue([]byte{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if val != "" {
			t.Errorf("val = %q, want empty", val)
		}
		if rest != nil {
			t.Errorf("rest = %q, want nil", rest)
		}
	})

	t.Run("unquoted basic", func(t *testing.T) {
		val, _, err := extractValue([]byte("value"), map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if val != "value" {
			t.Errorf("val = %q, want %q", val, "value")
		}
	})

	t.Run("unquoted with inline comment", func(t *testing.T) {
		val, _, err := extractValue([]byte("value # comment"), map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if val != "value" {
			t.Errorf("val = %q, want %q", val, "value")
		}
	})

	t.Run("unquoted no space before hash keeps hash", func(t *testing.T) {
		val, _, err := extractValue([]byte("value#notacomment"), map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if val != "value#notacomment" {
			t.Errorf("val = %q, want %q", val, "value#notacomment")
		}
	})

	t.Run("unquoted trimmed", func(t *testing.T) {
		val, _, err := extractValue([]byte("  value  "), map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if val != "value" {
			t.Errorf("val = %q, want %q", val, "value")
		}
	})

	t.Run("single-quoted", func(t *testing.T) {
		val, rest, err := extractValue([]byte("'value'\n"), map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if val != "value" {
			t.Errorf("val = %q, want %q", val, "value")
		}
		if string(rest) != "\n" {
			t.Errorf("rest = %q, want %q", string(rest), "\n")
		}
	})

	t.Run("double-quoted with expansion", func(t *testing.T) {
		val, _, err := extractValue([]byte(`"$A/world"`), map[string]string{"A": "hello"})
		if err != nil {
			t.Fatal(err)
		}
		if val != "hello/world" {
			t.Errorf("val = %q, want %q", val, "hello/world")
		}
	})

	t.Run("double-quoted with escapes", func(t *testing.T) {
		val, _, err := extractValue([]byte(`"\"quoted\""`), map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if val != `"quoted"` {
			t.Errorf("val = %q, want %q", val, `"quoted"`)
		}
	})
}

func TestProcessEscapes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`\"`, `"`},
		{`\\`, `\`},
		{`\$`, `\$`},
		{`\n`, "\n"},
		{`\r`, "\r"},
		{`\x`, `x`},
		{`"bar"`, `"bar"`},
	}
	for _, tt := range tests {
		got := processEscapes(tt.input)
		if got != tt.expected {
			t.Errorf("processEscapes(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestResolveVars(t *testing.T) {
	tests := []struct {
		input    string
		vars     map[string]string
		expected string
	}{
		{"$FOO", map[string]string{"FOO": "bar"}, "bar"},
		{"${FOO}", map[string]string{"FOO": "bar"}, "bar"},
		{"$FOO/world", map[string]string{"FOO": "hello"}, "hello/world"},
		{"$NOTHING", map[string]string{}, ""},
		{"\\$FOO", map[string]string{"FOO": "bar"}, "$FOO"},
		{"$(whoami)", map[string]string{}, "(whoami)"},
		{"$foo", map[string]string{"foo": "bar"}, "$foo"},
		{"hello", map[string]string{}, "hello"},
	}
	for _, tt := range tests {
		got := resolveVars(tt.input, tt.vars)
		if got != tt.expected {
			t.Errorf("resolveVars(%q, %v) = %q, want %q", tt.input, tt.vars, got, tt.expected)
		}
	}
}

func TestIsSpace(t *testing.T) {
	spaceRunes := []rune{'\t', '\v', '\f', '\r', ' ', 0x85, 0xA0}
	for _, r := range spaceRunes {
		if !isSpace(r) {
			t.Errorf("isSpace(%q) should be true", r)
		}
	}
	if isSpace('\n') {
		t.Error("isSpace('\\n') should be false")
	}
	if isSpace('a') {
		t.Error("isSpace('a') should be false")
	}
}

func TestIsLineEnd(t *testing.T) {
	if !isLineEnd('\n') {
		t.Error("isLineEnd('\\n') should be true")
	}
	if !isLineEnd('\r') {
		t.Error("isLineEnd('\\r') should be true")
	}
	if isLineEnd(' ') {
		t.Error("isLineEnd(' ') should be false")
	}
}

func TestHasQuotePrefix(t *testing.T) {
	if prefix, ok := hasQuotePrefix([]byte("'value")); !ok || prefix != '\'' {
		t.Errorf("expected single quote prefix")
	}
	if prefix, ok := hasQuotePrefix([]byte(`"value`)); !ok || prefix != '"' {
		t.Errorf("expected double quote prefix")
	}
	if _, ok := hasQuotePrefix([]byte("value")); ok {
		t.Errorf("expected no quote prefix")
	}
	if _, ok := hasQuotePrefix([]byte{}); ok {
		t.Errorf("expected no quote prefix on empty")
	}
}
