package models

import (
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type AuthDetails struct {
	USER_ID          int    `json:"user_id"`
	ROLE_ID          int    `json:"role_id"`
	USER_NAME        string `json:"user_name"`
	EMAIL            string `json:"email"`
	TENANT_ID        string `json:"tenant_id"`
	BUSINESS_ID      string `json:"business_id"`
	BUSINESS_UNIT_ID string `json:"business_unit_id"`
	REPORTING_ID     int    `json:"reporting_id"`
	UserHierarchy    int    `gorm:"column:user_hierarchy"`
}

type CustomClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	jwt.StandardClaims
}

func GetUsersListFromHeader(db *gorm.DB, subId string) (AuthDetails, error) {

	var users AuthDetails
	result := db.Table("ZNNXT_USER_MST").Select("USER_ID as user_id, ROLE_ID as role_id,EMAIL as email,TENANT_ID as tenant_id,BUSINESS_ID as business_id,BUSINESS_UNIT_ID as business_unit_id").Where("KEYCLOAK_ID = ?", subId).Scan(&users)
	if result.Error != nil {
		return users, result.Error
	}
	return users, nil
}
