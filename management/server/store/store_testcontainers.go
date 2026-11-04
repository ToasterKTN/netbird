//go:build testcontainers

package store

import "github.com/netbirdio/netbird/management/server/testutil"

func init() {
	createPostgresTestContainer = func() (string, error) {
		_, dsn, err := testutil.CreatePostgresTestContainer()
		return dsn, err
	}
}
