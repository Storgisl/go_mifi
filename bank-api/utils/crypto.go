package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func HashCVV(cvv string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
    return string(hash), err
}

func CheckCVV(cvv, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(cvv)) == nil
}
