package b2

import (
	"encoding/json"
	"fmt"
	"time"
)

// Session holds the information obtained from the login call.
//
// The information is sufficient enough to interact with the API
// directly. Expired sessions can contain attributes with zero
// values, so be sure to check whether the session has not expired
// before using it.
type Session struct {
	authorizationResponse

	// Timestamp of when the token will become invalid
	TokenExpiresAt time.Time `json:"tokenExpiresAt"`
}

// Expired returns whether the session has expired.
func (s *Session) Expired() bool {
	return timeNow().After(s.TokenExpiresAt)
}

// restoreSessionFromCache attempts to restore the session from cache.
//
// The cache might be disk based or an in-memory cache.
func restoreSessionFromCache(cache Cache) (*Session, error) {
	val, err := cache.Get("session")
	if err != nil {
		return nil, err
	}

	// Session cache does not exist, but this is not an error
	if val == nil {
		return &Session{}, nil
	}

	sessionBytes, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	var s *Session
	if err := json.Unmarshal(sessionBytes, &s); err != nil {
		return nil, fmt.Errorf("cannot restore session from cache: %v", val)
	}

	return s, nil
}

// commitSessionToCache attempts to persist the session in cache.
//
// The cache might be disk based or an in-memory cache.
func commitSessionToCache(cache Cache, session *Session) error {
	return cache.Set("session", session)
}
