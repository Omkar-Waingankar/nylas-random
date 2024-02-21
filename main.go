package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	grantID = "nylas_test_8"
	apiKey  = "REPLACE_WITH_API_KEY"
	limit   = 20
	maxIter = 10
	folder  = "DRAFT"

	local    = true
	provider = "google"
	email    = "nylas_test_8@nylas.com"
)

type Thread struct {
	ID                   string `json:"id"`
	LatestDraftOrMessage struct {
		Date int64 `json:"date"`
	} `json:"latest_draft_or_message"`
}

type ThreadsResponse struct {
	Data       []Thread `json:"data"`
	NextCursor string   `json:"next_cursor"`
}

func main() {
	// Make API request to fetch threads
	threads, err := fetchThreads()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Fetched %v threads\n", len(threads))

	// Check if threads are in descending order
	if !isDescendingOrder(threads) {
		fmt.Println("Error: Threads are not in descending order. Exiting...")
		return
	}
}

func fetchThreads() ([]Thread, error) {
	// Create an HTTP client
	client := &http.Client{}

	// Create a new request
	var u *url.URL
	var err error
	if local {
		u, err = url.Parse("http://localhost:6060/v3/threads")
		if err != nil {
			return nil, err
		}
	} else {
		u, err = url.Parse(fmt.Sprintf("https://api-staging.us.nylas.com/v3/grants/%s/threads", grantID))
		if err != nil {
			return nil, err
		}
	}

	q := u.Query()

	if limit != 0 {
		q.Set("limit", fmt.Sprintf("%v", limit))
	}

	if folder != "" {
		q.Set("in", folder)
	}

	u.RawQuery = q.Encode()

	fmt.Println("Request URL:", u.String())

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the necessary headers for authentication
	if local {
		req.Header.Set("X-Nylas-Provider-gma", provider)
		req.Header.Set("X-Nylas-Email-Address", email)
		req.Header.Set("X-Nylas-Grant-ID", grantID)
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}

	var allThreads []Thread
	pageToken := ""
	iteration := 0

	for iteration < maxIter {
		// Set the page token if available
		if pageToken != "" {
			q := req.URL.Query()
			q.Add("page_token", pageToken)
			req.URL.RawQuery = q.Encode()
		}

		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		fmt.Println("Response Status:", resp.Status)

		// Parse the response body into ThreadsResponse struct
		var threadsResponse ThreadsResponse
		err = json.NewDecoder(resp.Body).Decode(&threadsResponse)
		if err != nil {
			return nil, err
		}

		// Append the threads to the result
		allThreads = append(allThreads, threadsResponse.Data...)

		// Check if there are more pages to fetch
		if threadsResponse.NextCursor == "" {
			break
		}

		// Update the page token for the next iteration
		pageToken = threadsResponse.NextCursor
		iteration++
	}

	return allThreads, nil
}

func isDescendingOrder(threads []Thread) bool {
	inOrder := true

	// Check if threads are in descending order
	for i := 1; i < len(threads); i++ {
		if threads[i].LatestDraftOrMessage.Date > threads[i-1].LatestDraftOrMessage.Date {
			fmt.Printf("Error: Threads are not in descending order. Thread %v (index %v) is newer than thread %v (index %v)\n", threads[i].ID, i, threads[i-1].ID, i-1)
			inOrder = false
		}
	}

	if inOrder {
		fmt.Println("Threads are in descending order")
	}

	return inOrder
}
