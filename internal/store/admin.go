package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Admin struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	PasswordHash string    `json:"-" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Store) InitAdmin() (string, error) {
	var count int64
	if err := s.db.Model(&Admin{}).Count(&count).Error; err != nil {
		return "", fmt.Errorf("count admin: %w", err)
	}

	if count > 0 {
		return "", nil
	}

	password, err := generateRandomPassword(16)
	if err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	admin := &Admin{
		ID:           1,
		PasswordHash: string(hash),
	}

	if err := s.db.Create(admin).Error; err != nil {
		return "", fmt.Errorf("create admin: %w", err)
	}

	log.Println("=================================================")
	log.Printf("Admin password: %s", password)
	log.Println("Please change it after first login!")
	log.Println("=================================================")

	return password, nil
}

func (s *Store) VerifyPassword(password string) (bool, error) {
	var admin Admin
	if err := s.db.First(&admin, 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password))
	return err == nil, nil
}

func (s *Store) UpdatePassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.db.Model(&Admin{}).Where("id = ?", 1).Update("password_hash", string(hash)).Error
}

func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
