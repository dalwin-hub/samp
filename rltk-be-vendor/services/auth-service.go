package services

import (
	"fmt"
	"net/http"

	"github.com/Nerzal/gocloak/v12"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/gorilla/mux"
)

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

var Client *gocloak.GoCloak

var keyCloakClientID string
var keyCloakClientSecret string
var keyCloakRealm string

func InitializeOauthServer(host, clientId, clientSecret, realm string) {
	keyCloakClientID = clientId
	keyCloakClientSecret = clientSecret
	keyCloakRealm = realm

	Client = gocloak.NewClient(host)
}

func GenToken(name string, password string, r *http.Request) (*gocloak.JWT, error) {
	grantType := "password"
	jwt, err := Client.GetToken(r.Context(), keyCloakRealm, gocloak.TokenOptions{
		ClientID:     &keyCloakClientID,
		ClientSecret: &keyCloakClientSecret,
		Username:     &name,
		Password:     &password,
		GrantType:    &grantType,
	})

	if err != nil {
		fmt.Println("Error while creating Token for User", err)
		return nil, err
	}

	return jwt, nil
}

func RefreshToken(refreshToken string, r *http.Request) (*gocloak.JWT, error) {
	jwt, err := Client.RefreshToken(r.Context(), refreshToken, keyCloakClientID, keyCloakClientSecret, keyCloakRealm)
	if err != nil {
		fmt.Println("Error while creating Token for User", err)
		return nil, err
	}

	return jwt, nil
}

func Validate(accessToken string, r *http.Request) (*jwt.MapClaims, error) {
	claims, err := DecodeToken(accessToken, r)
	if err != nil {
		fmt.Println("Error while decoding the Token for User", err)
		return nil, err
	}

	_, err = Client.RetrospectToken(r.Context(), accessToken, keyCloakClientID, keyCloakClientSecret, keyCloakRealm)
	if err != nil {
		fmt.Println("Error while Inspecting the Token for User", err)
		return nil, err
	}

	return claims, nil
}

func DecodeToken(accessToken string, r *http.Request) (*jwt.MapClaims, error) {
	_, claims, err := Client.DecodeAccessToken(r.Context(), accessToken, keyCloakRealm)
	if err != nil {
		fmt.Println("Error while decoding the Token for User", err)
		return nil, err
	}

	return claims, nil
}
