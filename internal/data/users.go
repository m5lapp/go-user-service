package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/m5lapp/go-user-service/serialisation/jsonz"
	"github.com/m5lapp/go-user-service/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	AnonymousUser     = &User{}
	ErrDuplicateEmail = errors.New("duplicate email")
)

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 14)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

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

type User struct {
	ID           int64           `json:"id"`
	Version      int             `json:"-"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Email        string          `json:"email"`
	Password     password        `json:"-"`
	Name         string          `json:"name"`
	FriendlyName *string         `json:"friendly_name,omitempty"`
	BirthDate    *jsonz.DateOnly `json:"birth_date,omitempty"`
	Gender       *string         `json:"gender,omitempty"`
	CountryCode  *string         `json:"country_code,omitempty"`
	TimeZone     *string         `json:"time_zone,omitempty"`
	Activated    bool            `json:"activated"`
	Suspended    bool            `json:"suspended"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}

	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	if user.FriendlyName != nil {
		l := len(*user.FriendlyName) <= 500
		v.Check(l, "friendly_name", "must not be more than 500 bytes long")
	}

	if user.BirthDate != nil {
		tooOld := user.BirthDate.After(time.Now().AddDate(-120, 0, 0))
		v.Check(tooOld, "birth_date", "Must not be more than 120 years ago")
		tooYoung := user.BirthDate.Before(time.Now().AddDate(-13, 0, 0))
		v.Check(tooYoung, "birth_date", "Must be more than 13 years ago")
	}

	if user.Gender != nil {
		l := len(*user.Gender) <= 64
		v.Check(l, "gender", "must not be more than 64 bytes long")
	}

	if user.CountryCode != nil {
		// TODO: Ensure the country code is a valid option.
		v.Check(len(*user.CountryCode) == 2, "country_code", "must be exactly two bytes long")
	}

	if user.TimeZone != nil {
		_, err := time.LoadLocation(*user.TimeZone)
		v.Check(err == nil, "time_zone", "must be a valid time zone name")
	}
}

type UserModel struct {
	DB *sql.DB
}

// Insert adds the given User into the database. If the email address (case
// insensitive) already exists in the database, then an ErrDuplicateEmail
// response will be returned.
func (m UserModel) Insert(user *User) error {
	// The INSERT query returns the automatically generated values so that they
	// can be added to the User struct.
	query := `
		insert into users (
			email, password_hash, name, friendly_name, birth_date, gender,
			country_code, time_zone
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	 returning id, version, created_at, updated_at, activated, suspended
	`

	args := []any{
		user.Email,
		user.Password.hash,
		user.Name,
		user.FriendlyName,
		user.BirthDate,
		user.Gender,
		user.CountryCode,
		user.TimeZone,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, query, args...)
	err := row.Scan(&user.ID, &user.Version, &user.CreatedAt, &user.UpdatedAt, &user.Activated, &user.Suspended)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// GetByEmail queries the database for a user with the given email address. If
// no matching record exists, ErrRecordNotFound is returned.
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		select
		    users.id, users.version, users.created_at, users.updated_at,
			users.email, users.password_hash, users.name, users.friendly_name,
			users.birth_date, users.gender, users.country_code, users.time_zone,
		    users.activated, users.suspended
		  from users
		 where email = $1
		   and deleted = false
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Email,
		&user.Password.hash,
		&user.Name,
		&user.FriendlyName,
		&user.BirthDate,
		&user.Gender,
		&user.CountryCode,
		&user.TimeZone,
		&user.Activated,
		&user.Suspended,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update updates the database record for the given User. If there is an edit
// conflict and the version number is not the expected one, then ErrEditConflict
// will be returned.
func (m UserModel) Update(user *User) error {
	query := `
		update users
		   set version = version + 1, updated_at = now(),
		       email = $1, password_hash = $2, name = $3, friendly_name = $4,
			   birth_date = $5, gender = $6, country_code = $7, time_zone = $8,
			   activated = $9, suspended = $10
		 where id = $11 and version = $12 and deleted = false
		 returning version, updated_at, activated, suspended
	`

	args := []any{
		user.Email,
		user.Password.hash,
		user.Name,
		user.FriendlyName,
		user.BirthDate,
		user.Gender,
		user.CountryCode,
		user.TimeZone,
		user.Activated,
		user.Suspended,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, query, args...)
	err := row.Scan(&user.Version, &user.UpdatedAt, &user.Activated, &user.Suspended)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violoates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// GetForToken retrieves a User from the database for a given Token. If the
// token is expired, or the user has been suspended or deleted, then an
// ErrRecordNotFound error is returned.
func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	query := `
		select users.id, users.version, users.created_at, users.updated_at,
			   users.email, users.password_hash, users.name, users.friendly_name,
			   users.birth_date, users.gender, users.country_code,
			   users.time_zone, users.activated, users.suspended
		  from users
	inner join tokens
	        on users.id = tokens.user_id
		 where tokens.hash = $1
		   and tokens.scope = $2
		   and tokens.expiry > $3
		   and users.suspended = false
		   and users.deleted = false
	`

	var user User
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	// Convert tokenHash ([32]byte) to a slice as pq does not support arrays.
	args := []any{tokenHash[:], tokenScope, time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Version,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Email,
		&user.Password.hash,
		&user.Name,
		&user.FriendlyName,
		&user.BirthDate,
		&user.Gender,
		&user.CountryCode,
		&user.TimeZone,
		&user.Activated,
		&user.Suspended,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// DeleteByEmail soft deletes the user with the given email address. If
// no matching record exists, ErrRecordNotFound is returned.
func (m UserModel) DeleteByEmail(email string) error {
	deleteTokens := `delete from tokens where user_id = $1`
	deleteUser := `
		update users
		   set version = version + 1, updated_at = now(), deleted = true
		 where id = $1
		   and deleted = false
	`

	// Get the user first, this allows us to check they exist and have not
	// already been deleted. It also gives us their user ID so we can delete
	// their tokens.
	user, err := m.GetByEmail(email)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Delete all of the user's tokens.
	_, err = m.DB.ExecContext(ctx, deleteTokens, user.ID)
	if err != nil {
		return err
	}

	// Soft-delete the user record.
	_, err = m.DB.ExecContext(ctx, deleteUser, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}
