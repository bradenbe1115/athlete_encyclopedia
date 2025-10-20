package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

var errFail = errors.New("operation failed")

type mockDBConnector struct {
	connectErr error
}

func (m *mockDBConnector) Connect(ctx context.Context) (*sql.DB, error) {
	if m.connectErr != nil {
		return nil, m.connectErr
	}

	return &sql.DB{}, nil
}

func TestNew(t *testing.T) {
	tt := []struct {
		desc          string
		dbConnector   DBConnector
		expectedError error
		shouldBeNil   bool
	}{
		{
			desc:          "successful connection returns APIHandler",
			dbConnector:   &mockDBConnector{},
			expectedError: nil,
		},
		{
			desc:          "connection error returns error",
			dbConnector:   &mockDBConnector{connectErr: errFail},
			expectedError: errFail,
		},
	}

	for _, test := range tt {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			_, err := New(ctx, test.dbConnector)

			if !errors.Is(err, test.expectedError) {
				t.Errorf("unexpected error - want: %v, got: %v", test.expectedError, err)
			}
		})
	}
}
