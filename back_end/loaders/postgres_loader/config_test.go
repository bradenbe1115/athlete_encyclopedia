package main

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadConfigFromFile(t *testing.T) {
	tt := []struct {
		desc             string
		config_file_path string
		expected         *JobConfig
		expectedError    error
	}{
		{
			desc:             "a valid config with all fields included",
			config_file_path: "testdata/valid_config.json",
			expected: &JobConfig{
				InputType:     "json",
				DestTableName: "a_destination_table",
				LoadMethod:    "append",
			},
			expectedError: nil,
		},
		{
			desc:             "invalid config",
			config_file_path: "testdata/invalid_config.json",
			expected:         nil,
			expectedError:    ErrInvalidConfig,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			result, err := ReadConfigFromFile(test.config_file_path)
			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.expected, result); diff != "" {
				t.Errorf("expected -, actual +:\n%s", diff)
			}

		})
	}
}
