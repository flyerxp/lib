package safe

import (
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
)

type JwtL struct {
	JwtOptions *JwtOptions
}

type JwtOptions struct {
	Key           []byte
	SigningMethod *jwt.SigningMethodHMAC
}
type Option func(opts *JwtOptions)

func loadJwtJwtOptions(opts *JwtOptions, options ...Option) *JwtOptions {
	for _, option := range options {
		option(opts)
	}
	return opts
}
func CreateJwt(key []byte, f ...Option) *JwtL {
	j := new(JwtL)
	j.JwtOptions = new(JwtOptions)
	j.JwtOptions.Key = key
	j.JwtOptions = loadJwtJwtOptions(j.JwtOptions, f...)
	if j.JwtOptions.SigningMethod == nil {
		j.JwtOptions.SigningMethod = jwt.SigningMethodHS512
	}
	return j
}
func (j *JwtL) Encode(data jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(j.JwtOptions.SigningMethod, data)
	return token.SignedString(j.JwtOptions.Key)
}
func (j *JwtL) Decode(tokenString string) (jwt.MapClaims, error) {
	token2, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return j.JwtOptions.Key, nil
	})
	if claims, ok := token2.Claims.(jwt.MapClaims); ok && token2.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
