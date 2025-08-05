package middleware

import (
	"FinanceTracker/gateway/pkg/utils"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func NewAuth(secretKey []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) < len("Bearer ") || authHeader[:7] != "Bearer " {
				utils.WriteError(w, "invalid authorization format", http.StatusUnauthorized)
				return
			}

			token := authHeader[len("Bearer "):]

			claims, err := verify(token, secretKey)
			if err != nil {
				utils.WriteError(w, "invalid token", http.StatusUnauthorized)
				return
			}
			sub, err := claims.GetSubject()
			if err != nil {
				utils.WriteError(w, "invalid subject in token", http.StatusBadRequest)
				return
			}
			userId, err := strconv.ParseInt(sub, 10, 64)
			if err != nil {
				utils.WriteError(w, "invalid subject in token", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r.WithContext(utils.WithUserID(r.Context(), userId)))
		})
	}
}

func verify(tokenString string, secret []byte) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, jwt.ErrTokenNotValidYet
	}
	return token.Claims, nil
}
