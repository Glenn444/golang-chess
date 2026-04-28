package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

//Generates random OTP codes using crypto/rand
//digits is the number of codes to return
func GenerateOTP(digits int) (string, error) {
    max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
    n, err := rand.Int(rand.Reader, max)
    if err != nil {
        return "", err
    }
    // Zero-pad to ensure correct length (e.g. 047291)
    return fmt.Sprintf("%0*d", digits, n), nil
}

func SignOtpCode(otp string, secret string)(string,error){
	
	hm := hmac.New(sha256.New,[]byte(secret))
	hm.Write([]byte(otp))
	otpHashBytes := hm.Sum(nil)

    otpHashString := hex.EncodeToString(otpHashBytes)
    return otpHashString,nil
}

//Use hmac.Equal to compare code bytes
func ConfirmOTP(otp,otpHashString,secret string)(bool,error){
   
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(otp))
    expectedMAC := mac.Sum(nil)
    

    decodedOTPHashByte,err := hex.DecodeString(otpHashString)
    if err != nil{
        return false,err
    }

    // Always use hmac.Equal for security
    return hmac.Equal(decodedOTPHashByte, expectedMAC),nil
}