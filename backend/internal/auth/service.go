package auth

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/darkline28/ski-resort/backend/internal/users"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user users.User) (int, error)
	GetByEmail(ctx context.Context, email string) (users.User, error)
}

type Service struct {
	repo      UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewService(repo UserRepository, jwtSecret []byte, tokenTTL time.Duration) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}

func (s *Service) generateToken(id int) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(id),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

type RegisterInput struct {
	Name     string
	Email    string
	Phone    string
	Password string
}

var ErrInvalidInput = errors.New("invalid input")
var ErrEmailAlreadyExists = errors.New("email already exists")

func (s *Service) Register(ctx context.Context, input RegisterInput) (users.User, string, error) {
	email := strings.TrimSpace(input.Email)
	name := strings.TrimSpace(input.Name)
	phone := strings.TrimSpace(input.Phone)

	if name == "" || email == "" ||
		phone == "" || strings.TrimSpace(input.Password) == "" {
		return users.User{}, "", ErrInvalidInput
	}

	resMail, errMail := mail.ParseAddress(email)
	if errMail != nil || resMail.Address != email {
		return users.User{}, "", ErrInvalidInput
	}

	nameLength := utf8.RuneCountInString(name)
	if nameLength < 2 || nameLength > 100 {
		return users.User{}, "", ErrInvalidInput
	}

	matched, errPhone := regexp.MatchString(`^\+7[0-9]{10}$`, phone)
	if errPhone != nil || !matched {
		return users.User{}, "", ErrInvalidInput
	}

	lenPassword := len(input.Password)
	if lenPassword < 8 || lenPassword > 72 {
		return users.User{}, "", ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return users.User{}, "", err
	}

	user := users.User{
		Name:         name,
		Email:        email,
		Phone:        phone,
		PasswordHash: string(hash),
		Role:         "client",
	}

	id, err := s.repo.Create(ctx, user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
			return users.User{}, "", ErrEmailAlreadyExists
		}
		return users.User{}, "", err
	}
	user.ID = id
	token, err := s.generateToken(user.ID)
	if err != nil {
		return users.User{}, "", err
	}
	return user, token, nil

}

var ErrInvalidEmailOrPassword = errors.New("invalid email or password")

func (s *Service) Login(ctx context.Context, email, password string) (users.User, string, error) {
	inputEmail := strings.TrimSpace(email)
	if inputEmail == "" || strings.TrimSpace(password) == "" {
		return users.User{}, "", ErrInvalidInput
	}

	resMail, errMail := mail.ParseAddress(inputEmail)
	if errMail != nil || resMail.Address != inputEmail {
		return users.User{}, "", ErrInvalidInput
	}

	user, err := s.repo.GetByEmail(ctx, inputEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users.User{}, "", ErrInvalidEmailOrPassword
		}
		return users.User{}, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return users.User{}, "", ErrInvalidEmailOrPassword
	}
	token, err := s.generateToken(user.ID)
	if err != nil {
		return users.User{}, "", err
	}

	return user, token, nil
}
