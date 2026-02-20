package domain

var SessionKey = "session_id"

type CtxKey string

type AuthResult struct {
	UserID       int64  `json:"user_id"`
	UserLogin    string `json:"user_login"`
	UserName     string `json:"user_name"`
	UserLastName string `json:"user_last_name"`
	SessionID    string `json:"session_id"`
}
