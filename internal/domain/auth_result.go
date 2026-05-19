package domain

var SessionKey = "session_id"

type AuthResult struct {
	UserID       UserID
	UserLogin    string
	UserName     string
	UserLastName string
	SessionID    SessionID
}
