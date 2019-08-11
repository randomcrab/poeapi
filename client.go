package poeapi

import (
	"net/http"
	"time"
)

const (
	// DefaultHost sets the hostname used by the client. This is set to
	// api.pathofexile.com, but may be overridden for testing purposes.
	DefaultHost = "api.pathofexile.com"

	// DefaultNinjaHost sets the hostname used to retrieve the latest stash
	// change ID from poe.ninja. This is set to "poe.ninja" but may be
	// overridden for testing purposes.
	DefaultNinjaHost = "poe.ninja"

	// DefaultRateLimit sets the rate limit for all endpoints except for the
	// stash endpoint. Most endpoints have a rate limit of 5 requests per
	// second. Tests performed with the ratetest program (cmd/ratetest) show
	// occasional failures at this rate, so we back down to 4 requests per
	// second by default to err on the side of caution.
	DefaultRateLimit = 4

	// DefaultStashRateLimit sets the rate limit for the stash endpoint.
	// The stash API has a rate limit of 1 request per second.
	DefaultStashRateLimit = 1

	// DefaultCacheSize sets the number of items which can be stores in the
	// in-memory LRU cache. A typical response from the public stashs API
	// is around 3MB. By default, allow around 50x3MB=150MB total cache memory
	// usage.
	DefaultCacheSize = 50

	// DefaultRequestTimeout sets the time to wait before canceling HTTP
	// requests. Some endpoits take over 1000ms to respond, so we use 2000ms as
	// as default.
	DefaultRequestTimeout = 2 * time.Second
)

// APIClient provides methods for interacting with the Path of Exile API.
type APIClient interface {
	GetLeagues(GetLeaguesOptions) ([]League, error)
	GetLeague(GetLeagueOptions) (League, error)
	GetLeagueRules() ([]LeagueRule, error)
	GetLeagueRule(GetLeagueRuleOptions) (LeagueRule, error)
	GetLadder(GetLadderOptions) (Ladder, error)
	GetPVPMatches(GetPVPMatchesOptions) ([]PVPMatch, error)
	GetStashes(opts GetStashOptions) (StashResponse, error)
	GetLatestStashID() (string, error)
}

type client struct {
	httpClient *http.Client

	host      string
	ninjaHost string

	useSSL   bool
	useCache bool

	limiter *ratelimiter
	cache   *cache
}

// NewAPIClient configures and returns an APIClient.
func NewAPIClient(opts ClientOptions) (APIClient, error) {
	if err := validateClientOptions(opts); err != nil {
		return nil, err
	}

	c := &client{
		host:       opts.Host,
		ninjaHost:  opts.NinjaHost,
		useSSL:     opts.UseSSL,
		useCache:   opts.UseCache,
		limiter:    newRateLimiter(opts.RateLimit, opts.StashRateLimit),
		httpClient: &http.Client{Timeout: opts.RequestTimeout},
	}

	if opts.UseCache {
		cache, err := newCache(opts.CacheSize)
		if err != nil {
			return nil, err
		}
		c.cache = cache
	}

	return c, nil
}

// ClientOptions contains settings for client initialization.
type ClientOptions struct {
	// The hostname used by the client.
	Host string

	// The hostname used for poe.ninja requests.
	NinjaHost string

	// Set to false if your network does not allow outbound HTTPS traffic.
	UseSSL bool

	// Set to false to always hit the API for requests instead of using a local
	// cache.
	UseCache bool

	// The number of items which can be stored in the cache. Expect around
	// 3MB per item for stash requests, and up to 0.5MB per item for all
	// other requests.
	CacheSize int

	// The number of requests per second for all API endpoints except the stash
	// tab endpoint. The API will ratelimit clients above 5rps.
	RateLimit int

	// The number of requests per second for the stash endpoint. The API
	// will ratelimit clients above 1rps.
	StashRateLimit int

	// Time to wait before canceling HTTP requests.
	RequestTimeout time.Duration
}

// DefaultClientOptions initializes the client with the most common settings.
var DefaultClientOptions = ClientOptions{
	Host:           DefaultHost,
	NinjaHost:      DefaultNinjaHost,
	UseSSL:         true,
	UseCache:       true,
	CacheSize:      DefaultCacheSize,
	RateLimit:      DefaultRateLimit,
	StashRateLimit: DefaultStashRateLimit,
	RequestTimeout: DefaultRequestTimeout,
}

func validateClientOptions(opts ClientOptions) error {
	if _, ok := validHosts[opts.Host]; !ok {
		return ErrInvalidHost
	}
	if opts.NinjaHost == "" {
		return ErrInvalidNinjaHost
	}
	if opts.CacheSize < 1 {
		return ErrInvalidCacheSize
	}
	if opts.RateLimit < 0 {
		return ErrInvalidRateLimit
	}
	if opts.StashRateLimit < 0 {
		return ErrInvalidStashRateLimit
	}
	if opts.RequestTimeout < 1*time.Millisecond {
		return ErrInvalidRequestTimeout
	}
	return nil
}
