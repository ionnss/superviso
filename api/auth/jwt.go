// Package auth fornece funcionalidades de autenticação e autorização.
//
// Implementa:
//   - Geração e validação de tokens JWT
//   - Middleware de autenticação
//   - Controle de contexto para usuário autenticado

// Claims define a estrutura do payload do JWT.
// Inclui ID e email do usuário, além dos claims padrão do JWT.
package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtKey é a chave secreta para assinar e verificar tokens JWT.
//
// A chave é obtida do ambiente ou definida como uma string fixa.
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// Claims define a estrutura do payload do JWT.
// Inclui ID e email do usuário, além dos claims padrão do JWT.
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ContextKey é um tipo personalizado para chaves de contexto
// para evitar colisões com outras chaves no contexto.
type ContextKey string

// Constantes para chaves de contexto.
// UserIDKey: chave para ID do usuário no contexto
// EmailKey: chave para email do usuário no contexto
const (
	UserIDKey ContextKey = "user_id"
	EmailKey  ContextKey = "email"
)

// GenerateToken cria um novo token JWT para um usuário.
//
// Parâmetros:
//   - userID: ID do usuário
//   - email: Email do usuário
//
// Retorna:
//   - string: token JWT assinado
//   - error: erro se houver falha na geração
func GenerateToken(userID int, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateToken verifica se o token JWT é válido e retorna suas claims.
//
// Parâmetros:
//   - tokenString: token JWT em formato string
//
// Retorna:
//   - *Claims: claims do token se válido
//   - error: erro se token inválido ou expirado
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	return claims, nil
}

// AuthMiddleware protege rotas que precisam de autenticação.
//
// Verifica se existe um token JWT válido no cookie e adiciona
// informações do usuário (ID e email) no contexto da requisição.
//
// Parâmetros:
//   - next: próxima função handler a ser executada
//
// Retorna:
//   - http.HandlerFunc que verifica autenticação
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			log.Printf("Erro ao obter cookie: %v", err)
			http.Error(w, "Não autorizado", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateToken(cookie.Value)
		if err != nil {
			log.Printf("Erro ao validar token: %v", err)
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		log.Printf("Token válido para usuário ID: %d, Email: %s", claims.UserID, claims.Email)

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
