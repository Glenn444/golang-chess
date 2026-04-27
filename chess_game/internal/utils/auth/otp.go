package auth

import (
    "crypto/rand"
    "fmt"
    "math/big"
)


//Generates random OTP codes using crypto/rand
//digits is the number of codes to return
func generateOTP(digits int) (string, error) {
    max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
    n, err := rand.Int(rand.Reader, max)
    if err != nil {
        return "", err
    }
    // Zero-pad to ensure correct length (e.g. 047291)
    return fmt.Sprintf("%0*d", digits, n), nil
}