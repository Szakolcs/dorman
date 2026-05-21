package cross_cutting

import (
	models "dorm-man/internal/models/cross-cutting"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte("secret-key")

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
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
	JWTToken, err := createToken(u)
	return LoginResponse{
		UserID:      u.ID.String(),
		Token:       JWTToken,
		Name:        u.Name,
		Email:       u.Email,
		RedirectTo:  u.Role.Name,
		LastLoginAt: time.Time{},
	}, nil
}

func createToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":       user.ID,
			"username": user.Name,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
			"role":     user.Role.Name,
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
