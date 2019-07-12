package installer

import (
	"testing"
)

func TestTrimWhitespace(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Nothing to trim", "Nothing to trim"},
		{"  \n\nOne Two \n ", "One Two"},
	}

	for _, testCase := range testCases {
		actual := trimWhitespace([]byte(testCase.input))

		if string(actual) != testCase.expected {
			t.Errorf("Expected string = %s; got string = %s", testCase.expected, actual)
		}
	}
}

func TestTrimShebang(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"#!/bin/bash\ncode", "code"},
		{"code\n#!/bin/bash\ncode", "code\n#!/bin/bash\ncode"},
	}

	for _, testCase := range testCases {
		actual := trimShebang([]byte(testCase.input))

		if string(actual) != testCase.expected {
			t.Errorf("Expected string = %s; got string = %s", testCase.expected, actual)
		}
	}
}

func TestExpandEnvironment(t *testing.T) {
	envMap := map[string]string{
		"VARIABLE": "myVariable",
	}

	testCases := []struct {
		input    string
		expected string
	}{
		{"Testing ${VARIABLE}", "Testing myVariable"},
		{"testing ${INVALID}", "testing "},
	}

	origEnvGetter := envGetter
	defer func() { envGetter = origEnvGetter }()

	envGetter = func(key string) string {
		return envMap[key]
	}

	for _, testCase := range testCases {
		actual := expandEnvironment([]byte(testCase.input))

		if string(actual) != testCase.expected {
			t.Errorf("Expected string = %s; got string = %s", testCase.expected, actual)
		}
	}
}
