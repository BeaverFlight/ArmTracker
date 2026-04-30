package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func LoadOrGeneratePrivateKey(path string, logger *slog.Logger) (*rsa.PrivateKey, error) {
	key, err := loadPrivateKey(path)
	if err == nil {
		logger.Info("RSA ключ загружен из файла", slog.String("path", path))
		return key, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("чтение RSA ключа: %w", err)
	}

	logger.Warn("файл RSA ключа не найден, генерируем новый", slog.String("path", path))

	key, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("генерация RSA ключа: %w", err)
	}

	if err := savePrivateKey(path, key); err != nil {
		logger.Error("не удалось сохранить RSA ключ на диск",
			slog.String("path", path), slog.String("error", err.Error()))
	} else {
		logger.Info("новый RSA ключ сохранён", slog.String("path", path))
	}

	return key, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("PEM блок не найден в файле ключа")
	}

	// Пробуем PKCS8, затем PKCS1
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("ключ в файле не является RSA")
		}
		return rsaKey, nil
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func savePrivateKey(path string, key *rsa.PrivateKey) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("создание директории для ключа: %w", err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("сериализация RSA ключа: %w", err)
	}

	block := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})

	return os.WriteFile(path, block, 0o600)
}
