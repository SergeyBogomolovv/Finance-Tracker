package dto

type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderYandex OAuthProvider = "yandex"
)

type OAuthPayload struct {
	Email     string
	FullName  string
	AvatarUrl string
	Provider  OAuthProvider
}
