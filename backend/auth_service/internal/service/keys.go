package service

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("чтение ключа: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("PEM блок не найден")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Пробуем PKCS1
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	return key.(*rsa.PrivateKey), nil
}
