package repository

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestNextQueuedClusterNoDuplicates simulates several cluster instances polling the same postgres database
// and verifies that every queued task is claimed exactly once.
// Requires a postgres database, e.g.:
//
//	docker run -d --rm -e POSTGRES_PASSWORD=pw -p 55499:5432 postgres:16
//	FFMATE_TEST_POSTGRES="postgres://postgres:pw@127.0.0.1:55499/postgres?sslmode=disable" go test ./internal/database/repository/
func TestNextQueuedClusterNoDuplicates(t *testing.T) {
	dsn := os.Getenv("FFMATE_TEST_POSTGRES")
	if dsn == "" {
		t.Skip("FFMATE_TEST_POSTGRES not set")
	}

	open := func() *gorm.DB {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatal(err)
		}
		return db
	}

	setup := open()
	_ = setup.Migrator().DropTable("task_labels", "client_labels", &model.Task{}, &model.Client{}, &model.Label{})
	(&Client{DB: setup}).Setup()
	(&Task{DB: setup}).Setup()

	const tasks, instances, perPoll = 300, 6, 8
	if err := setup.Create(&model.Client{Identifier: "creator"}).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < tasks; i++ {
		if err := setup.Create(&model.Task{UUID: uuid.NewString(), Name: fmt.Sprint(i), Status: dto.Queued, ClientIdentifier: "creator"}).Error; err != nil {
			t.Fatal(err)
		}
	}

	var mu sync.Mutex
	claims := map[uint]int{}
	var wg sync.WaitGroup
	for n := 0; n < instances; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			repo := &Task{DB: open()} // separate connection pool per instance
			for {
				got, err := repo.NextQueued(perPoll, nil)
				if err != nil {
					t.Error(err)
					return
				}
				if got == nil {
					return
				}
				mu.Lock()
				for _, task := range *got {
					claims[task.ID]++
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	assert.Len(t, claims, tasks, "every task must be claimed")
	for id, c := range claims {
		assert.Equal(t, 1, c, "task %d claimed %d times", id, c)
	}
}

// TestFailRunningTasksForOfflineClientsClusterOnce simulates several cluster instances cleaning up concurrently
// and verifies that every task of an offline client is failed (and returned for notification) exactly once.
func TestFailRunningTasksForOfflineClientsClusterOnce(t *testing.T) {
	dsn := os.Getenv("FFMATE_TEST_POSTGRES")
	if dsn == "" {
		t.Skip("FFMATE_TEST_POSTGRES not set")
	}
	open := func() *gorm.DB {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatal(err)
		}
		return db
	}

	setup := open()
	_ = setup.Migrator().DropTable("task_labels", "client_labels", &model.Task{}, &model.Client{}, &model.Label{})
	(&Client{DB: setup}).Setup()
	(&Task{DB: setup}).Setup()

	if err := setup.Create(&model.Client{Identifier: "offline", LastSeen: 1}).Error; err != nil {
		t.Fatal(err)
	}
	const tasks, instances = 300, 6
	statuses := []dto.TaskStatus{dto.PreProcessing, dto.Running, dto.PostProcessing}
	for i := 0; i < tasks; i++ {
		if err := setup.Create(&model.Task{UUID: uuid.NewString(), Status: statuses[i%3], ClientIdentifier: "offline"}).Error; err != nil {
			t.Fatal(err)
		}
	}

	var mu sync.Mutex
	returned := map[uint]int{}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for n := 0; n < instances; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			repo := &Task{DB: open()}
			<-start
			failed, err := repo.FailRunningTasksForOfflineClients()
			if err != nil {
				t.Error(err)
				return
			}
			mu.Lock()
			for _, task := range failed {
				returned[task.ID]++
			}
			mu.Unlock()
		}()
	}
	close(start)
	wg.Wait()

	assert.Len(t, returned, tasks, "every task must be failed")
	for id, c := range returned {
		assert.Equal(t, 1, c, "task %d returned %d times", id, c)
	}
}
