package cross_cutting

import (
	"dorm-man/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	store         Store
	sessionSecret []byte
}

func NewService(store Store, sessionSecret string) *Service {
	return &Service{
		store:         store,
		sessionSecret: []byte(sessionSecret),
	}
}

func (s *Service) Login(userCred LoginRequest) (LoginResponse, error) {
	u, err := s.store.GetUserByNickname(userCred.Nickname)
	if err != nil {
		return LoginResponse{}, err
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash),
		[]byte(userCred.Password),
	); err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}
	if err := s.store.UpdateLastLogin(u.ID, time.Now()); err != nil {
		return LoginResponse{}, ErrLoginUpdateFailed
	}
	JWTToken, err := s.createToken(u)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{
		UserID:      u.ID.String(),
		Token:       JWTToken,
		Name:        u.Name,
		Email:       u.Email,
		RedirectTo:  u.Role.Name,
		LastLoginAt: time.Time{},
	}, nil
}

func (s *Service) createToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":       user.ID.String(),
			"username": user.Name,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
			"role":     user.Role.Name,
		})

	tokenString, err := token.SignedString(s.sessionSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
