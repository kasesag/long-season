package users

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/hakierspejs/long-season/pkg/models"
	"github.com/hakierspejs/long-season/pkg/services/happier"
	"github.com/hakierspejs/long-season/pkg/storage"
	"golang.org/x/crypto/bcrypt"
)

const (
	invalidNicknameMsg = "username should contains from 4 to 32 numerical and alphabetical characters"
	invalidPasswordMsg = "password should contains from 6 to 50 any characters, excluding whitespace characters"
)

var (
	// ErrInvalidNickname error is used for various verifies to
	// signal that user verification failed because of
	// invalid username. Raw message of error is safe to output
	// to client.
	ErrInvalidNickname = errors.New(invalidNicknameMsg)

	// ErrInvaliPassword error is used for various verifies to
	// signal that user verification failed because of
	// invalid password. Raw message of error is safe to output
	// to client.
	ErrInvaliPassword = errors.New(invalidPasswordMsg)
)

var (
	nicknameRegex = regexp.MustCompile(`^[a-zA-Z0-9]{4,32}$`)
	passwordRegex = regexp.MustCompile(`^[^[:space:]]{6,50}$`)
)

// VerifyNickname verifies if given nickname string
// is proper nickname for long-season application.
func VerifyNickname(n string) bool {
	return nicknameRegex.MatchString(n)
}

// VerifyPassword verifies if given password string
// is proper password for long-season application.
func VerifyPassword(p string) bool {
	return passwordRegex.MatchString(p)
}

// VerifyRegisterData verifies if given data required for
// user registration is valid. Returned error messages are
// safe to output to client.
func VerifyRegisterData(nickname, password string) error {
	if ok := VerifyNickname(nickname); !ok {
		return ErrInvalidNickname
	}

	if ok := VerifyPassword(password); !ok {
		return ErrInvaliPassword
	}

	return nil
}

// AuthenticateRequest holds input data for AuthenticateWithPassword
// function.
type AuthenticateRequest struct {
	// Nickname is used to find user.
	Nickname string

	// Password is used to verify user with
	// requested nickname.
	Password []byte
}

// AuthenticateDependencies are dependencies for
// authenticating user with password.
type AuthenticateDependencies struct {
	// Request holds input data.
	Request AuthenticateRequest

	// Storage operates on user data in database.
	Storage storage.Users

	// ErrorFactory is optional. If you want
	// to have debug errors, you can pass
	// error Factory created from http request.
	ErrorFactory *happier.Factory
}

func AuthenticateWithPassword(ctx context.Context, deps AuthenticateDependencies) (*models.User, error) {
	if deps.ErrorFactory == nil {
		// ErrorFactory is optional parameter, so if it is nil
		// we replace it with default happier error factory.
		deps.ErrorFactory = happier.Default()
	}

	users, err := deps.Storage.All(ctx)
	if err != nil {
		return nil, deps.ErrorFactory.InternalServerError(
			fmt.Errorf("deps.Storage.All: %w", err),
			"Internal Server Error please try again later.",
		)
	}

	// Search for user with exactly same nickname.
	var match *models.User = nil
	for _, user := range users {
		if user.Nickname == deps.Request.Nickname {
			match = &user
			break
		}
	}

	// Check if there is the user with given nickname
	// in the database.
	if match == nil {
		return nil, deps.ErrorFactory.NotFound(
			fmt.Errorf("match == nil, user given nickname: %s, not found", deps.Request.Nickname),
			fmt.Sprintf("there is no user with given nickname: \"%s\"", deps.Request.Nickname),
		)
	}

	// Check if passwords do match.
	if err := bcrypt.CompareHashAndPassword(
		match.Password,
		deps.Request.Password,
	); err != nil {
		return nil, deps.ErrorFactory.Unauthorized(
			fmt.Errorf("bcrypt.CompareHashAndPassword: %w", err),
			fmt.Sprintf("given password does not match"),
		)
	}

	return match, nil
}
