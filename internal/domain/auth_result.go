package domain

var SessionKey = "session_id"

type CtxKey string

type AuthResult struct {
	UserID       UserID
	UserLogin    string
	UserName     string
	UserLastName string
	SessionID    SessionID
}
