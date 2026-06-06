package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestMakeAndValidateJWT verifica o caminho feliz: cria e valida com sucesso.
func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "meu-segredo-super-secreto"

	tokenString, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT falhou: %v", err)
	}

	gotID, err := ValidateJWT(tokenString, secret)
	if err != nil {
		t.Fatalf("ValidateJWT falhou: %v", err)
	}

	if gotID != userID {
		t.Errorf("esperava userID %v, obteve %v", userID, gotID)
	}
}

// TestExpiredToken verifica que tokens expirados são rejeitados.
func TestExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "meu-segredo"

	// Duração negativa → já nasce expirado
	tokenString, err := MakeJWT(userID, secret, -1*time.Second)
	if err != nil {
		t.Fatalf("MakeJWT falhou: %v", err)
	}

	_, err = ValidateJWT(tokenString, secret)
	if err == nil {
		t.Fatal("esperava erro para token expirado, mas não obteve nenhum")
	}
}

// TestWrongSecret verifica que tokens assinados com segredo errado são rejeitados.
func TestWrongSecret(t *testing.T) {
	userID := uuid.New()

	tokenString, err := MakeJWT(userID, "segredo-correto", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT falhou: %v", err)
	}

	_, err = ValidateJWT(tokenString, "segredo-errado")
	if err == nil {
		t.Fatal("esperava erro para segredo errado, mas não obteve nenhum")
	}
}