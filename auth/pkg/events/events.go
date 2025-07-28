package events

const (
	TopicRegistered   = "user.registered"
	TopicOTPGenerated = "user.otp.generated"
)

type EventOTPGenerated struct {
	UserID int    `json:"user_id"`
	Code   string `json:"otp"`
}

type EventUserRegistered struct {
	UserID int `json:"user_id"`
}
