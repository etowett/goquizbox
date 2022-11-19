package auth

import (
	"encoding/json"
	"os"
	"time"

	"goquizbox/internal/entities"
	"goquizbox/internal/repo/model"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/text/language"
)

type (
	JWTHandler interface {
		CreateUserToken(*model.User, *model.Session) (string, error)
		TokenInfo(tokenValue string) (*entities.TokenInfo, error)
	}

	AppJWTHandler struct {
		keyFunc    func(*jwt.Token) (interface{}, error)
		signingKey []byte
	}
)

func NewJWTHandler() JWTHandler {
	return NewJWTHandlerWithSigningKey(os.Getenv("JWT_SIGNING_KEY"))
}

func NewJWTHandlerWithSigningKey(signingKey string) JWTHandler {

	handler := &AppJWTHandler{
		signingKey: []byte(signingKey),
	}

	handler.keyFunc = func(*jwt.Token) (interface{}, error) {
		return handler.signingKey, nil
	}

	return handler
}

func (h *AppJWTHandler) CreateUserToken(
	user *model.User,
	session *model.Session,
) (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)

	claims := make(jwt.MapClaims, 0)

	claims["exp"] = time.Now().AddDate(1, 0, 0).Unix()
	claims["refresh"] = time.Now().Add(time.Hour).Unix()
	claims["session_id"] = session.ID
	claims["status"] = user.Status.String()
	claims["user_id"] = user.ID

	token.Claims = claims

	return token.SignedString(h.signingKey)
}

func (h *AppJWTHandler) TokenInfo(tokenValue string) (*entities.TokenInfo, error) {

	token, err := jwt.Parse(tokenValue, h.keyFunc)
	if err != nil {
		return &entities.TokenInfo{}, err
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &entities.TokenInfo{}, nil
	}

	exp := h.getInt64(mapClaims, "exp")
	refresh := h.getInt64(mapClaims, "refresh")
	sessionID := h.getInt64(mapClaims, "session_id")
	status := h.getString(mapClaims, "status")
	userID := h.getInt64(mapClaims, "user_id")

	tokenInfo := &entities.TokenInfo{
		Exp:       time.Unix(exp, 0),
		Refresh:   time.Unix(refresh, 0),
		SessionID: sessionID,
		Status:    status,
		UserID:    userID,
	}

	return tokenInfo, nil
}

func (h *AppJWTHandler) getInt64(mapClaims jwt.MapClaims, key string) int64 {
	switch val := mapClaims[key].(type) {
	case float64:
		return int64(val)
	case json.Number:
		v, _ := val.Int64()
		return v
	}
	return 0
}

func (h *AppJWTHandler) getString(mapClaims jwt.MapClaims, key string) string {
	switch val := mapClaims[key].(type) {
	case string:
		return string(val)
	}
	return ""
}

func (h *AppJWTHandler) getBool(mapClaims jwt.MapClaims, key string) bool {
	switch val := mapClaims[key].(type) {
	case bool:
		return bool(val)
	}

	return false
}

func (h *AppJWTHandler) getLang(mapClaims jwt.MapClaims, key string) string {
	switch val := mapClaims[key].(type) {
	case string:
		return string(val)
	}
	return language.AmericanEnglish.String()
}
