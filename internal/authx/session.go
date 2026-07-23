package authx

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const sessionName = "carnet_session"
const userIDKey = "user_id"

type Sessions struct {
	store *sessions.CookieStore
}

func NewSessions(secretKey string) *Sessions {
	store := sessions.NewCookieStore([]byte(secretKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30, // 30 jours
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return &Sessions{store: store}
}

func (s *Sessions) Login(w http.ResponseWriter, r *http.Request, userID int32) error {
	sess, _ := s.store.Get(r, sessionName)
	sess.Values[userIDKey] = int(userID)
	return sess.Save(r, w)
}

func (s *Sessions) Logout(w http.ResponseWriter, r *http.Request) error {
	sess, _ := s.store.Get(r, sessionName)
	sess.Values[userIDKey] = nil
	sess.Options.MaxAge = -1
	return sess.Save(r, w)
}

// CurrentUserID returns the logged-in user's id, or 0 if not authenticated.
func (s *Sessions) CurrentUserID(r *http.Request) int32 {
	sess, err := s.store.Get(r, sessionName)
	if err != nil {
		return 0
	}
	v, ok := sess.Values[userIDKey].(int)
	if !ok {
		return 0
	}
	return int32(v)
}

func (s *Sessions) IsAuthenticated(r *http.Request) bool {
	return s.CurrentUserID(r) != 0
}
