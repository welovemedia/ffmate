package repository

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/testsuite/testserver"
)

// every sqlite connection to :memory: gets its own empty database, so concurrent operations
// must all run on the single connection that has the tables
func TestInMemoryDatabaseParallelOperations(t *testing.T) {
	server := testserver.New(t)
	repo := (&Preset{DB: server.DB()}).Setup()

	sqlDB, err := server.DB().DB()
	assert.NoError(t, err)
	assert.Equal(t, 1, sqlDB.Stats().MaxOpenConnections)

	const n = 50
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.Save(&model.Preset{UUID: uuid.NewString(), Name: "p", Command: "-y"})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		assert.NoError(t, err)
	}

	count, err := repo.Count()
	assert.NoError(t, err)
	assert.Equal(t, int64(n), count, "all writes must land in the same database")
}
