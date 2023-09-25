package env

import (
	"os"
	"reflect"
	"testing"
)

var values = map[string]string{
	"APP_LOGLEVEL": "TRACE",
	"APP_PORT":     "12345",
	"APP_ENABLED":  "True",

	"APP_HOMESERVER": "https://example.com",
	"APP_LOGIN":      "@test:example.com",
	"APP_PASSWORD":   "password",

	"APP_SPAM_EMAILS": "ima@spammer.com definetelynotspam@gmail.com",
	"APP_SPAM_HOSTS":  "spamer.com unitedspammers.org",

	"APP_BAN_DURATION": "1",
	"APP_BAN_SIZE":     "invalid",

	"APP_LIST": "test1 test2",

	"APP_BOOL_TRUE":   "True",
	"APP_BOOL_YES":    "yEs",
	"APP_BOOL_TRUEUP": "TRUE",

	"APP_TEST1_REDIRECT":  "https://example.org",
	"APP_TEST1_RATELIMIT": "1r/s",
	"APP_TEST1_ROOM":      "!test1@example.com",

	"APP_TEST2_REDIRECT":  "https://example.com",
	"APP_TEST2_RATELIMIT": "1r/m",
	"APP_TEST2_ROOM":      "!test2@example.com",
}

func TestMain(m *testing.M) {
	for key, value := range values {
		os.Setenv(key, value)
	}

	SetPrefix("app")
	exit := m.Run()

	for key := range values {
		os.Unsetenv(key)
	}
	os.Exit(exit)
}

func TestString(t *testing.T) {
	tests := []struct {
		name         string
		shortkey     string
		defaultValue string
		expected     string
	}{
		{"exists", "login", "", "@test:example.com"},
		{"notexists_with_default", "none", "default", "default"},
		{"notexists", "none", "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual string
			if test.defaultValue == "" {
				actual = String(test.shortkey)
			} else {
				actual = String(test.shortkey, test.defaultValue)
			}

			if test.expected != actual {
				t.Error(test.expected, "is not", actual)
			}
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		name         string
		shortkey     string
		defaultValue int
		expected     int
	}{
		{"exists", "ban.duration", 0, 1},
		{"notexists_with_default", "none", 5, 5},
		{"invalid", "ban.size", 0, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Int(test.shortkey, test.defaultValue)

			if test.expected != actual {
				t.Error(test.expected, "is not", actual)
			}
		})
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		name     string
		shortkey string
		expected bool
	}{
		{"exists", "enabled", true},
		{"notexists", "none", false},
		{"invalid", "ban.size", false},
		{"true", "bool.true", true},
		{"trueup", "bool.trueup", true},
		{"yes", "bool.yes", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Bool(test.shortkey)

			if test.expected != actual {
				t.Error(test.expected, "is not", actual)
			}
		})
	}
}

func TestSlice(t *testing.T) {
	tests := []struct {
		name     string
		shortkey string
		expected []string
	}{
		{"exists", "SPAM_HOSTS", []string{"spamer.com", "unitedspammers.org"}},
		{"one", "ban.size", []string{"invalid"}},
		{"notexists", "none", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Slice(test.shortkey)

			if !reflect.DeepEqual(test.expected, actual) {
				t.Error(test.expected, "is not", actual)
			}
		})
	}
}
