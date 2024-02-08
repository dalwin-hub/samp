package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rltk-be-vendor/services"
	"rltk-be-vendor/utils"
	"strings"
)

// SetMiddlewareJSON is used to set the content-type header
func SetMiddlewareJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r)

		if r := recover(); r != nil {
			// Log the error
			utils.GetLogger().WithError(fmt.Errorf("%v", r)).Error("In middlewares.go line 21,Error occurred in SetMiddlewareJSON")

			// Respond with an error to the client if needed
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// SetMiddlewareAuthentication make sures the incoming request is authorised
// with JWT token
func SetMiddlewareAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if len(authHeader) < 1 {
			utils.GetLogger().Error("In middlewares.go line 36,Authorization header is missing")
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(utils.UnauthorizedError())
			return
		}

		accessToken := authHeader
		if len(strings.Split(authHeader, " ")) == 2 {
			accessToken = strings.Split(authHeader, " ")[1]
		}

		claims, err := services.Validate(accessToken, r)
		if err != nil {
			utils.GetLogger().WithError(err).Error("In middlewares.go line 49,Error validating access token")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(utils.BadRequestError(err.Error()))
			return
		}

		vErr := claims.Valid()
		if vErr != nil {
			utils.GetLogger().Error("In middlewares.go line 57,Invalid token claims")
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(utils.UnauthorizedError())
			return
		}
		//create a new request context containing the authenticated user
		ctxWithUser := context.WithValue(r.Context(), "user", *claims)
		//create a new request using that new context
		rWithUser := r.WithContext(ctxWithUser)

		next(w, rWithUser)
	}
}
