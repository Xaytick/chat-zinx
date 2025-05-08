package protocol

type TextMsg struct {
	ToUserID uint32 `json:"to_user_id"`
	Content  string `json:"content"`
}
