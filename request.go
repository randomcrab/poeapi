package poeapi

import (
	"io/ioutil"
	"net/http"
)

// getJSON retrieves the given URL, checking the local cache before making an
// external request. It returns the JSON response as a string.
func (c *client) getJSON(url string) (string, error) {
	if c.useCache {
		cached := c.cache.Get(url)
		if cached != "" {
			return cached, nil
		}
	}

	ratelimit := c.limiter.rateLimit
	if url == c.formatURL(stashTabsEndpoint) {
		ratelimit = c.limiter.stashTabRateLimit
	}
	c.limiter.wait(ratelimit)

	resp, err := http.Get(url)
	if err != nil {
		// An error is returned if the Client's CheckRedirect function fails or
		// if there was an HTTP protocol error. A non-2xx response doesn't cause
		// an error.
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// Continue.
	case http.StatusBadRequest:
		return "", ErrBadRequest
	case http.StatusNotFound:
		return "", ErrNotFound
	case http.StatusTooManyRequests:
		return "", ErrRateLimited
	case http.StatusInternalServerError:
		return "", ErrServerFailure
	default:
		return "", ErrUnknownFailure
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	body := string(b)
	if c.useCache {
		c.cache.Add(url, body)
	}

	return body, nil
}
