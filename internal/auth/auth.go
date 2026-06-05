package auth

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) string{
	res, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil{
		return ""
	}
	return res
}

func CheckPasswordHash(password, hash string) (bool, error){
	res, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil{
		return false, err
	}
	return res, nil
}