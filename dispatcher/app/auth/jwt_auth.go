package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type Service struct {
	Opts
}

type Opts struct {
	JWTHeader   string
}

type Claims struct {
	jwt.StandardClaims
	TenantID int `json:"tid,omitempty"`
	ClientID int `json:"oid,omitempty"`
	AppID    string `json:"azp,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
}

const (
	defaultJWTHeader   = "Authorization"
)

func NewService(opts Opts) *Service {
	res := Service{Opts: opts}

	setDefault := func(opt *string, defValue string) {
		if len(*opt) == 0 {
			*opt = defValue
		}
	}

	setDefault(&res.JWTHeader, defaultJWTHeader)

	return &res
}

func (s *Service) Parse(tokenStr string) (Claims, error) {
	parser := jwt.Parser{SkipClaimsValidation: true}
	token, err := parser.ParseWithClaims(tokenStr, &Claims{} , func(token *jwt.Token) (interface{}, error) {
		return []byte("your-256-bit-secret"), nil
	})
	if err != nil {
		return Claims{}, errors.Wrap(err, "can't parse token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return Claims{}, errors.New("invalid token")
	}

	return *claims, s.validate(claims)
}

func (s *Service) validate(claims *Claims) error {
	err := claims.Valid()

	if err != nil {
		return err
	}

	return nil
}