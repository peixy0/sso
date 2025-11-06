package login

import (
	"os"

	"golang.org/x/crypto/bcrypt"
)

type Authenticator struct {
	encryptedMasterKey []byte
}

func NewAuthenticator() *Authenticator {
	key := os.Getenv("SSO_MASTERKEY")
	return &Authenticator{
		encryptedMasterKey: []byte(key),
	}
}

func (a *Authenticator) Authenticate(secret string) bool {
	err := bcrypt.CompareHashAndPassword(a.encryptedMasterKey, []byte(secret))
	return err == nil
}
