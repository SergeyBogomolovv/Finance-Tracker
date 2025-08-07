package events

const (
	TopicRegistered = "user.registered"
)

type EventUserRegistered struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	Provider  string `json:"provider"`
	FullName  string `json:"full_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}
