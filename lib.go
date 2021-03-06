package nanoapisession

import (
	"context"
	"errors"
	"net/http"

	"github.com/digitalcircle-com-br/nanoapi"
	. "github.com/digitalcircle-com-br/nanoapi-log"
)

type Session struct {
	ID        string
	User      string
	Perms     map[string]string
	Tenant    string
	ExtraInfo map[string]string
}

var SessionSave func(c context.Context, s Session) error
var SessionLoad func(c context.Context, id string) (*Session, error)
var SessionExist func(c context.Context, id string) (bool, error)
var SessionDel func(c context.Context, id string) error

var ErrSessionNotFound = errors.New("session not found")

//CtxSessionID - gets session id from req cookie.
func CtxSessionID(c context.Context) string {
	f := nanoapi.CtxReq(c)
	if f == nil {
		return ""
	}
	v := f.Header.Get("X-SESSION")
	if v != "" {
		return v
	}
	ck, err := f.Cookie("SESSION")

	if err == nil {
		if ck.Value != "" {
			return ck.Value
		}
	}
	if f.URL.Query().Get("session") != "" {
		return f.URL.Query().Get("session")
	}
	return ""
}

//CtxSessionExist - checks if session exists.
func CtxSessionExist(c context.Context) bool {
	id := CtxSessionID(c)
	if id == "" {
		return false
	}
	ret, err := SessionExist(c, id)
	if err != nil {
		Err("CtxSessionExist::error %s", err.Error())
		return false
	}
	return ret
}

//CtxSession - gets all session data from cache.
func CtxSession(c context.Context) *Session {
	id := CtxSessionID(c)
	if id == "" {
		return nil
	}
	ret, err := SessionLoad(c, id)
	if err != nil {
		Err("CtxSession::error %s", err.Error())
		return nil
	}
	return ret
}

//CtxMustSession - gets session and returns ErrSessionNotFound if something goes wrong.
func CtxMustSession(c context.Context) (*Session, error) {
	id := CtxSessionID(c)
	if id == "" {
		return nil, ErrSessionNotFound
	}
	ret, err := SessionLoad(c, id)
	if err != nil {
		Err("CtxSession::error %s", err.Error())
		return nil, ErrSessionNotFound
	}
	return ret, nil
}

//CtxSession - gets all session data from cache.
func ReqSession(r *http.Request) *Session {
	maybeSession := r.Context().Value("SESSION")
	if maybeSession != nil {
		sess, ok := maybeSession.(*Session)
		if ok {
			return sess
		}
	}
	ck, err := r.Cookie("SESSION")
	if err != nil {
		return nil
	}

	if ck.Value == "" {
		return nil
	}
	ret, err := SessionLoad(r.Context(), ck.Value)
	if err != nil {
		return nil
	}
	ctx := context.WithValue(r.Context(), "SESSION", ret)
	r.WithContext(ctx)
	return ret
}

func Setup() error {

	SessionExist = func(c context.Context, id string) (bool, error) {
		return true, nil
	}

	SessionLoad = func(c context.Context, id string) (*Session, error) {
		return nil, nil
	}

	SessionSave = func(c context.Context, s Session) error {
		return nil
	}

	nanoapi.CheckPerm = func(ctx context.Context, p string) bool {
		s := CtxSession(ctx)
		if s == nil {
			return false
		}
		if p == nanoapi.PERM_AUTH {
			return true
		}
		_, ok := s.Perms[p]
		return ok
	}

	return nil
}
