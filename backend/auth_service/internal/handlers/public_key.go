package handlers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func (h *Handlers) GetPublicKey(c *gin.Context) {
	pub := h.srv.PublicKey()
	jwk := rsaToJWK(pub, "auth-service-key-v1")
	c.JSON(http.StatusOK, JWKS{Keys: []JWK{jwk}})
}

func (h *Handlers) GetPublicKeyPEM(c *gin.Context) {
	pub := h.srv.PublicKey()

	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сериализовать ключ"})
		return
	}

	block := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	})

	c.Data(http.StatusOK, "application/x-pem-file", block)
}

func rsaToJWK(pub *rsa.PublicKey, kid string) JWK {
	nBytes := pub.N.Bytes()

	e := big.NewInt(int64(pub.E))
	eBytes := e.Bytes()

	return JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(nBytes),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}
}
