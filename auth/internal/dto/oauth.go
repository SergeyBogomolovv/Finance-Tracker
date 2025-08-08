package dto

const (
	OAuthProviderGoogle = "google"
	OAuthProviderYandex = "yandex"
)

type OAuthPayload struct {
	Email     string
	FullName  string
	AvatarUrl string
	Provider  string
}
