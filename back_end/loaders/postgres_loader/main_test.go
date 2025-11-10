package main

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtractImportIdFromFileName(t *testing.T) {
	tt := []struct {
		desc          string
		rawFilePath   string
		expected      string
		expectedError error
	}{
		{
			desc:          "valid file path",
			rawFilePath:   `athlete_encyclopedia/data/players/college_football/players_08-11-2025-22-0820.json`,
			expected:      `08/11/2025/22/0820`,
			expectedError: nil,
		},

		{
			desc:          "Not a file path",
			rawFilePath:   `athlete_encyclopedia/data/players/college_football/players_08-11-2025-22-0820`,
			expected:      `08/11/2025/22/0820`,
			expectedError: nil,
		},

		{
			desc:          "Invalid file format",
			rawFilePath:   `25-22-0820`,
			expected:      ``,
			expectedError: ErrUnexpectedFileFormat,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			result, err := ExtractImportIdFromFileName(test.rawFilePath)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.expected, result); diff != "" {
				t.Errorf("expected -, actual +:\n%s", diff)
			}
		})
	}
}
