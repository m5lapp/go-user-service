package data

import (
	"errors"

	"github.com/m5lapp/go-service-toolkit/validator"
	"golang.org/x/crypto/bcrypt"
)

// The password struct represents a password.
type password struct {
	plaintext *string
	hash      []byte
}

// Set generates the hash of the provided plaintextPassword and sets the
// plaintext and hash values of the password struct to the appropriate values.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 14)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches compares a plaintext password with its hash.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// ValidatePasswordPlaintext ensures that a provided password satisfies the
// desired password requirements. Any violations will be added to the given
// validator.Validator under the "password" key.
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
