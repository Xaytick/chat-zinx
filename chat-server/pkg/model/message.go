package model

type TextMsg struct {
	ToUserID string `json:"to_user_id"`
	Content  string `json:"content"`
}
