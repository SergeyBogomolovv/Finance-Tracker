package events

import "time"

const (
	TopicRegistered   = "user.registered"
	TopicOTPGenerated = "user.otp.generated"
)

type EventOTPGenerated struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type EventUserRegistered struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	Provider  string `json:"provider"`
	FullName  string `json:"full_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}
