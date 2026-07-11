package token

import (
	"testing"
	"time"

	"github.com/Glenn444/golang-chess/internal/utils/random"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)


func TestJWTMaker(t *testing.T){
	m, err := NewJWTMaker(random.RandomString(32))
	require.NoError(t,err)

	username := random.RandomUsername()
	duration := time.Minute

	issuedAt := time.Now()
	expireAt := time.Now().Add(duration)

	token,errToken := m.CreateToken(username,AccessTokenType,duration)
	require.NoError(t,errToken)
	require.NotEmpty(t,token)

	payload,errPayload := m.VerifyToken(token,AccessTokenType)
	require.NoError(t,errPayload)
	require.NotEmpty(t,payload)

	require.NotZero(t,payload.ID)
	require.Equal(t,username,payload.Username)
	require.WithinDuration(t,issuedAt,payload.IssuedAt.Time,time.Second)
	require.WithinDuration(t,expireAt,payload.ExpiresAt.Time,time.Second)

}

func TestExpiredToken(t *testing.T){
	m, err := NewJWTMaker(random.RandomString(32))
	require.NoError(t,err)

	username := random.RandomUsername()
	duration := -time.Second

	token,err := m.CreateToken(username,AccessTokenType,duration)
	require.NoError(t,err)
	require.NotEmpty(t,token)

	payload,errPayload := m.VerifyToken(token,AccessTokenType)
	require.Error(t,errPayload)
	require.EqualError(t,errPayload,ErrExpiredToken.Error())
	require.Nil(t,payload)

}


func TestInvalidJWTTokenAlgNone(t *testing.T){
	// 1. Create the payload
	payload,err := NewPayload(random.RandomUsername(),AccessTokenType,time.Minute)
	require.NoError(t,err)
	require.NotEmpty(t,payload)

	//2. create the token
	jwttoken := jwt.NewWithClaims(jwt.SigningMethodNone,payload)
	token, err := jwttoken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NotEmpty(t,token)
	require.NoError(t,err)


	//3. create a new maker
	m,err := NewJWTMaker(random.RandomString(32))

	//4. verify the token
	payloadToken,err := m.VerifyToken(token,AccessTokenType)
	require.Error(t,err)
	require.Nil(t,payloadToken)
	require.EqualError(t,err,ErrInvalidToken.Error())
}