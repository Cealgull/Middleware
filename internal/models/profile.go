package models

type UserProfile struct {
	UserId    string `json:"userId"`
	Username  string `json:"username,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Signature string `json:"signature,omitempty"`
}
