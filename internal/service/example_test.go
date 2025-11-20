package service_test

import (
	"fmt"
	"sort"

	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/Aleksey170999/go-shortener/internal/service"
)

// ExampleURLService_GetUserURLs demonstrates how to sort URLs by creation time.
// This example will be included in the package documentation.
func ExampleURLService_GetUserURLs() {
	// Create a new in-memory repository
	repo := repository.NewMemoryURLRepository()

	// Initialize the URL service
	urlService := service.NewURLService(repo)

	// Create a test user
	userID := "user123"

	// Create some test URLs
	urls := []model.URL{
		{ID: "1", Original: "https://example.com/1", Short: "abc123", UserID: userID},
		{ID: "2", Original: "https://example.com/2", Short: "def456", UserID: userID},
		{ID: "3", Original: "https://example.com/3", Short: "ghi789", UserID: userID},
	}

	// Save the URLs to the repository
	for _, u := range urls {
		url := u // Create a copy to avoid referencing the loop variable
		_, _ = repo.Save(&url)
	}

	// Retrieve and sort the user's URLs
	userURLs, _ := urlService.GetUserURLs(userID)

	// Sort the URLs by ID (which represents creation time)
	sort.Slice(userURLs, func(i, j int) bool {
		return userURLs[i].ID < userURLs[j].ID
	})

	// Print the sorted URLs
	for _, url := range userURLs {
		fmt.Printf("%s -> %s\n", url.Short, url.Original)
	}

	// Output:
	// abc123 -> https://example.com/1
	// def456 -> https://example.com/2
	// ghi789 -> https://example.com/3
}
