// Copyright 2018 The go-xcosh Authors
// This file is part of the go-xcosh library.
//
// The go-xcosh library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-xcosh library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-xcosh library. If not, see <http://www.gnu.org/licenses/>.

package accounts

import (
	"testing"
)

func TestURLParsing(t *testing.T) {
	t.Parallel()
	url, err := parseURL("https://xcosh.org")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("expected: %v, got: %v", "https", url.Scheme)
	}
	if url.Path != "xcosh.org" {
		t.Errorf("expected: %v, got: %v", "xcosh.org", url.Path)
	}

	for _, u := range []string{"xcosh.org", ""} {
		if _, err = parseURL(u); err == nil {
			t.Errorf("input %v, expected err, got: nil", u)
		}
	}
}

func TestURLString(t *testing.T) {
	t.Parallel()
	url := URL{Scheme: "https", Path: "xcosh.org"}
	if url.String() != "https://xcosh.org" {
		t.Errorf("expected: %v, got: %v", "https://xcosh.org", url.String())
	}

	url = URL{Scheme: "", Path: "xcosh.org"}
	if url.String() != "xcosh.org" {
		t.Errorf("expected: %v, got: %v", "xcosh.org", url.String())
	}
}

func TestURLMarshalJSON(t *testing.T) {
	t.Parallel()
	url := URL{Scheme: "https", Path: "xcosh.org"}
	json, err := url.MarshalJSON()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(json) != "\"https://xcosh.org\"" {
		t.Errorf("expected: %v, got: %v", "\"https://xcosh.org\"", string(json))
	}
}

func TestURLUnmarshalJSON(t *testing.T) {
	t.Parallel()
	url := &URL{}
	err := url.UnmarshalJSON([]byte("\"https://xcosh.org\""))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url.Scheme != "https" {
		t.Errorf("expected: %v, got: %v", "https", url.Scheme)
	}
	if url.Path != "xcosh.org" {
		t.Errorf("expected: %v, got: %v", "https", url.Path)
	}
}

func TestURLComparison(t *testing.T) {
	t.Parallel()
	tests := []struct {
		urlA   URL
		urlB   URL
		expect int
	}{
		{URL{"https", "xcosh.org"}, URL{"https", "xcosh.org"}, 0},
		{URL{"http", "xcosh.org"}, URL{"https", "xcosh.org"}, -1},
		{URL{"https", "xcosh.org/a"}, URL{"https", "xcosh.org"}, 1},
		{URL{"https", "abc.org"}, URL{"https", "xcosh.org"}, -1},
	}

	for i, tt := range tests {
		result := tt.urlA.Cmp(tt.urlB)
		if result != tt.expect {
			t.Errorf("test %d: cmp mismatch: expected: %d, got: %d", i, tt.expect, result)
		}
	}
}
