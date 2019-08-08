package poeapi

import "time"

// League represents a permanent or challenge league.
type League struct {
	Name         string       `json:"id"`
	Realm        string       `json:"realm"`
	Description  string       `json:"description"`
	LadderURL    string       `json:"url"`
	StartTime    time.Time    `json:"startAt"`
	EndTime      time.Time    `json:"endAt"`
	DelveEnabled bool         `json:"delveEvent"`
	Rules        []LeagueRule `json:"rules"`
}

// LeagueRule represents a modifier placed on a league.
type LeagueRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
