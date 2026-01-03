package jsonmessage

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/moby/moby/api/types/jsonstream"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestProgressString(t *testing.T) {
	type expected struct {
		short string
		long  string
	}

	shortAndLong := func(short, long string) expected {
		return expected{short: short, long: long}
	}

	start := time.Date(2017, 12, 3, 15, 10, 1, 0, time.UTC)
	timeAfter := func(delta time.Duration) func() time.Time {
		return func() time.Time {
			return start.Add(delta)
		}
	}

	testcases := []struct {
		name     string
		progress jsonstream.Progress
		expected expected
		nowFunc  func() time.Time
	}{
		{
			name: "no progress",
		},
		{
			name:     "progress 1",
			progress: jsonstream.Progress{Current: 1},
			expected: shortAndLong("      1B", "      1B"),
		},
		{
			name: "some progress with a start time",
			progress: jsonstream.Progress{
				Current: 20,
				Total:   100,
				Start:   start.Unix(),
			},
			nowFunc: timeAfter(time.Second),
			expected: shortAndLong(
				"     20B/100B 4s",
				"[==========>                                        ]      20B/100B 4s",
			),
		},
		{
			name:     "some progress without a start time",
			progress: jsonstream.Progress{Current: 50, Total: 100},
			expected: shortAndLong(
				"     50B/100B",
				"[=========================>                         ]      50B/100B",
			),
		},
		{
			name:     "current more than total is not negative gh#7136",
			progress: jsonstream.Progress{Current: 50, Total: 40},
			expected: shortAndLong(
				"     50B",
				"[==================================================>]      50B",
			),
		},
		{
			name:     "with units",
			progress: jsonstream.Progress{Current: 50, Total: 100, Units: "units"},
			expected: shortAndLong(
				"50/100 units",
				"[=========================>                         ] 50/100 units",
			),
		},
		{
			name:     "current more than total with units is not negative ",
			progress: jsonstream.Progress{Current: 50, Total: 40, Units: "units"},
			expected: shortAndLong(
				"50 units",
				"[==================================================>] 50 units",
			),
		},
		{
			name:     "hide counts",
			progress: jsonstream.Progress{Current: 50, Total: 100, HideCounts: true},
			expected: shortAndLong(
				"",
				"[=========================>                         ] ",
			),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.nowFunc != nil {
				originalTimeNow := timeNow
				timeNow = testcase.nowFunc
				defer func() { timeNow = originalTimeNow }()
			}
			assert.Equal(t, RenderTUIProgress(testcase.progress, 100), testcase.expected.short)
			assert.Equal(t, RenderTUIProgress(testcase.progress, 200), testcase.expected.long)
		})
	}
}

func TestJSONMessageDisplay(t *testing.T) {
	messages := map[jsonstream.Message][]string{
		// Empty
		{}: {"\n", "\n"},
		// Status
		{
			Status: "status",
		}: {
			"status\n",
			"status\n",
		},
		// General
		{
			ID:     "ID",
			Status: "status",
		}: {
			"ID: status\n",
			"ID: status\n",
		},
		// Stream over status
		{
			Status: "status",
			Stream: "stream",
		}: {
			"stream",
			"stream",
		},
		// With progress, stream empty
		{
			Status:   "status",
			Stream:   "",
			Progress: &jsonstream.Progress{Current: 1},
		}: {
			"",
			fmt.Sprintf("%c[2K\rstatus       1B\r", 27),
		},
	}

	// The tests :)
	for jsonMessage, expectedMessages := range messages {
		// Without terminal
		data := bytes.NewBuffer([]byte{})
		if err := Display(jsonMessage, data, false, 0); err != nil {
			t.Fatal(err)
		}
		if data.String() != expectedMessages[0] {
			t.Fatalf("Expected %q,got %q", expectedMessages[0], data.String())
		}
		// With terminal
		data = bytes.NewBuffer([]byte{})
		if err := Display(jsonMessage, data, true, 0); err != nil {
			t.Fatal(err)
		}
		if data.String() != expectedMessages[1] {
			t.Fatalf("\nExpected %q\n     got %q", expectedMessages[1], data.String())
		}
	}
}

// Test JSONMessage with an Error. It returns an error with the given text, not the meaning of the HTTP code.
func TestJSONMessageDisplayWithJSONError(t *testing.T) {
	data := bytes.NewBuffer([]byte{})
	jsonMessage := jsonstream.Message{Error: &jsonstream.Error{Code: 404, Message: "Can't find it"}}

	err := Display(jsonMessage, data, true, 0)
	if err == nil || err.Error() != "Can't find it" {
		t.Fatalf("Expected a jsonstream.Error 404, got %q", err)
	}

	jsonMessage = jsonstream.Message{Error: &jsonstream.Error{Code: 401, Message: "Anything"}}
	err = Display(jsonMessage, data, true, 0)
	assert.Check(t, is.Error(err, "Anything"))
}
