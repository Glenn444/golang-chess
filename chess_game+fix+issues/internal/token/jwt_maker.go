package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


const minSecretKey = 32

type JWTMaker struct{
	secretKey string
}

func NewJWTMaker(secretKey string)(Maker,error){
	if len(secretKey) < minSecretKey{
		return nil,fmt.Errorf("invalid key size: must be atleast %d characters\n",minSecretKey)
	}
	return &JWTMaker{
		secretKey: secretKey,
	},nil
}

// creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(username string,tokenType TokenType,duration time.Duration) (string, error){
	//steps
	// 1. create the token payload
	tokenPayload,err := NewPayload(username,tokenType,duration)
	if err != nil{
		return "",err
	}

	//2. create the JWT Token with NewWithClaims method
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256,tokenPayload)
	

	signedJwtTokenString, err := jwtToken.SignedString([]byte(maker.secretKey))
	return signedJwtTokenString,nil
}

	//verifyToken checks if the token is valid or not
func (maker *JWTMaker) VerifyToken(token string,tokenType TokenType) (*Payload, error){
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		_,ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok{
			return nil,ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	jwtToken, err := jwt.ParseWithClaims(token,&Payload{},keyFunc)
	if err != nil{
		if errors.Is(err,jwt.ErrTokenExpired){
			return nil,ErrExpiredToken
		}else{
		return nil,ErrInvalidToken
		}
	}

	//2. convert the JWT token into a payload
	payloadToken,ok := jwtToken.Claims.(*Payload)
	if !ok || payloadToken.TokenTpe != tokenType{
		return nil,ErrInvalidToken
	}
	return payloadToken,nil

}

