package utils

import (
    "crypto/rand"
    "fmt"
    "math/big"
)

func generateOTP(digits int) (string, error) {
    max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
    n, err := rand.Int(rand.Reader, max)
    if err != nil {
        return "", err
    }
    // Zero-pad to ensure correct length (e.g. 047291)
    return fmt.Sprintf("%0*d", digits, n), nil
}