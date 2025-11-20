package repository_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryURLRepository(t *testing.T) {
	repo := repository.NewMemoryURLRepository()
	testURL := &model.URL{
		ID:        "test1",
		Short:     "test1",
		Original:  "https://example.com",
		UserID:    "user1",
		IsDeleted: false,
	}

	t.Run("Save and GetByShortURL", func(t *testing.T) {
		savedURL, err := repo.Save(testURL)
		require.NoError(t, err)
		assert.Equal(t, testURL, savedURL)

		foundURL, err := repo.GetByShortURL(testURL.ID)
		require.NoError(t, err)
		assert.Equal(t, testURL, foundURL)

		_, err = repo.GetByShortURL("nonexistent")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("GetByUserID", func(t *testing.T) {
		// Save the test URL first
		savedURL, err := repo.Save(testURL)
		require.NoError(t, err)

		// Now try to get it by user ID
		urls, err := repo.GetByUserID("user1")
		require.NoError(t, err)
		require.Len(t, urls, 1)
		assert.Equal(t, savedURL.Short, urls[0].Short)

		// Test with non-existent user
		_, err = repo.GetByUserID("nonexistent")
		require.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("BatchDelete", func(t *testing.T) {
		err := repo.BatchDelete([]string{testURL.ID}, "user1")
		require.NoError(t, err)

		url, err := repo.GetByShortURL(testURL.ID)
		require.NoError(t, err)
		assert.True(t, url.IsDeleted)

		err = repo.BatchDelete([]string{"nonexistent"}, "user1")
		require.NoError(t, err)
	})

	// Add these test cases to the existing test file

	t.Run("Save duplicate URL", func(t *testing.T) {
		// First save should succeed
		_, err := repo.Save(testURL)
		require.NoError(t, err)

		// Second save with same ID should also succeed in memory implementation
		// (database implementation would return error)
		_, err = repo.Save(testURL)
		require.NoError(t, err)
	})

	t.Run("BatchDelete empty slice", func(t *testing.T) {
		err := repo.BatchDelete([]string{}, "user1")
		require.NoError(t, err)
	})

	t.Run("Concurrent access", func(t *testing.T) {
		const numWorkers = 10
		const urlsPerWorker = 100

		var wg sync.WaitGroup
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < urlsPerWorker; j++ {
					url := &model.URL{
						ID:        fmt.Sprintf("url-%d-%d", workerID, j),
						Short:     fmt.Sprintf("short-%d-%d", workerID, j),
						Original:  fmt.Sprintf("https://example.com/%d/%d", workerID, j),
						UserID:    fmt.Sprintf("user-%d", workerID%3), // Distribute across 3 users
						IsDeleted: false,
					}
					_, err := repo.Save(url)
					require.NoError(t, err)
				}
			}(i)
		}

		wg.Wait()

		// Verify all URLs were saved correctly
		for i := 0; i < numWorkers; i++ {
			for j := 0; j < urlsPerWorker; j++ {
				shortURL := fmt.Sprintf("short-%d-%d", i, j)
				url, err := repo.GetByShortURL(shortURL)
				require.NoError(t, err)
				require.Equal(t, shortURL, url.Short)
			}
		}

		// Test concurrent deletes
		wg.Add(3) // 3 users
		for i := 0; i < 3; i++ {
			go func(userNum int) {
				defer wg.Done()
				userID := fmt.Sprintf("user-%d", userNum)
				urls, err := repo.GetByUserID(userID)
				require.NoError(t, err)

				var ids []string
				for _, url := range urls {
					ids = append(ids, url.Short)
				}

				err = repo.BatchDelete(ids, userID)
				require.NoError(t, err)
			}(i)
		}
		wg.Wait()
	})
}
