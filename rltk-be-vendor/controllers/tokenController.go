package controllers

import (
	"errors"
	"net/http"
	"rltk-be-vendor/db"
	"rltk-be-vendor/models"
	"rltk-be-vendor/utils"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func GetUsers(w http.ResponseWriter, r *http.Request) (models.AuthDetails, error) {

	user := models.AuthDetails{}

	var token string

	bearertoken := r.Header.Get("Authorization")
	parts := strings.Split(bearertoken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		//fmt.Println("Invalid token format")
		utils.GetLogger().Error("In token controller line 24,Invalid token format")
		return user, errors.New("invalid token format")
	}
	if parts != nil {
		token = parts[1]
	} else {
		return user, nil
	}

	customClaims := models.CustomClaims{}
	token12, _, err := new(jwt.Parser).ParseUnverified(token, &customClaims)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In token controller line 36,Error occurred while parsing token")
		return user, err
	}

	claims, ok := token12.Claims.(*models.CustomClaims)
	if !ok {
		return user, nil
	}

	userDetails, err := GetUserFromToken(w, r, claims.Sub)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In token controller line 47,Error occurred while getting user details from token")
		return user, err
	} else {
		return userDetails, nil
	}
}

func GetUserFromToken(w http.ResponseWriter, r *http.Request, subId string) (models.AuthDetails, error) {
	//get the authenticated user from the request context
	users, err := models.GetUsersListFromHeader(db.GetDB(), subId)
	if err != nil {
		utils.GetLogger().WithError(err).Error("In token controller line 58,Error occurred while getting user details from database")
		utils.ERROR(w, http.StatusInternalServerError, err)
		return users, err
	}
	return users, nil
}
