package envfile

import "testing"

func TestParseQuoting(t *testing.T) {
	tests := []struct {
		name string
		src  string
		key  string
		want string
	}{
		{"double quoted", "A=\"bar baz\"\n", "A", "bar baz"},
		{"single quoted", "A='bar baz'\n", "A", "bar baz"},
		{"backslashes kept in single quotes", "A='C:\\Temp\\logs'\n", "A", `C:\Temp\logs`},
		{"backslashes kept unquoted", "A=C:\\Temp\\logs\n", "A", `C:\Temp\logs`},
		{"unknown escapes kept in double quotes", "A=\"C:\\Temp\\logs\"\n", "A", `C:\Temp\logs`},
		{"regex kept in double quotes", "A=\"\\d+\\s*\"\n", "A", `\d+\s*`},
		{"escaped backslash", "A=\"foo\\\\\"\n", "A", `foo\`},
		{"newline escape", "A=\"line1\\nline2\"\n", "A", "line1\nline2"},
		{"tab escape", "A=\"col1\\tcol2\"\n", "A", "col1\tcol2"},
		{"carriage return escape", "A=\"a\\rb\"\n", "A", "a\rb"},
		{"escaped quotes", "A=\"say \\\"hi\\\"\"\n", "A", `say "hi"`},
		{"escaped dollar", "A=\"cost \\$5\"\n", "A", "cost $5"},
		{"escaped dollar unquoted", "A=cost \\$5\n", "A", "cost $5"},
		{"single quotes stay raw", "A='raw \\n $FOO # x'\n", "A", `raw \n $FOO # x`},
		{"brace without dollar", "A=\"value {x}\"\n", "A", "value {x}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := parse(t, tt.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out[tt.key] != tt.want {
				t.Errorf("%s = %q, want %q", tt.key, out[tt.key], tt.want)
			}
		})
	}
}

func TestParseExpansion(t *testing.T) {
	tests := []struct {
		name string
		src  string
		key  string
		want string
	}{
		{"reference earlier key", "A=hello\nB=$A/world\n", "B", "hello/world"},
		{"braced reference", "A=hello\nB=${A}!\n", "B", "hello!"},
		{"lowercase reference", "a=1\nB=$a/x\n", "B", "1/x"},
		{"later key is invisible", "B=$A/x\nA=hello\n", "B", "$A/x"},
		{"self reference stays literal", "A=$A\n", "A", "$A"},
		{"missing reference stays literal", "A=$MISSING/x\n", "A", "$MISSING/x"},
		{"braced missing reference stays literal", "A=template ${VAR} here\n", "A", "template ${VAR} here"},
		{"unclosed brace stays literal", "A=${FOO\n", "A", "${FOO"},
		{"stray closing brace is kept", "A=$A}bar\n", "A", "$A}bar"},
		{"dollar digits stay literal", "A=x$123\n", "A", "x$123"},
		{"trailing dollar stays literal", "A=50$\n", "A", "50$"},
		{"password with undefined reference stays intact", "A=p@ss$WORD123\n", "A", "p@ss$WORD123"},
		{"bcrypt hash with undefined references stays intact", "A=$2a$10$N9qo8uLOickgx2ZMRZoMye\n", "A", "$2a$10$N9qo8uLOickgx2ZMRZoMye"},
		{"password loses its suffix to a defined reference", "WORD123=overridden\nA=p@ss$WORD123\n", "A", "p@ssoverridden"},
		{"expansion inside double quotes", "A=first\nB=\"$A x\"\n", "B", "first x"},
		{"escaped reference unquoted", "A=\\$HOME\n", "A", "$HOME"},
		{"escaped reference quoted", "A=\"\\$HOME\"\n", "A", "$HOME"},
		{"reference inside single quotes", "A='$FOO'\n", "A", "$FOO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := parse(t, tt.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out[tt.key] != tt.want {
				t.Errorf("%s = %q, want %q", tt.key, out[tt.key], tt.want)
			}
		})
	}
}
