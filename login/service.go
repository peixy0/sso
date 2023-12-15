package login

import (
	"os"

	"golang.org/x/crypto/bcrypt"
)

type LoginService struct {
	encryptedMasterKey string
}

func NewLoginService() *LoginService {
	key := os.Getenv("SSO_MASTERKEY")
	return &LoginService{
		encryptedMasterKey: key,
	}
}

func (s *LoginService) IsValidCredential(secret string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(s.encryptedMasterKey), []byte(secret))
	return err == nil
}
