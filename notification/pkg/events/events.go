package events

const (
	AuthOTPGeneratedTopic = "auth.otp.generated"
)

type OTPGeneratedEvent struct {
	UserID int    `json:"user_id"`
	Code   string `json:"otp_code"`
}
