package db

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Wallet struct {
	UserID          int64   `bson:"UserID"`
	TokenType       string  `bson:"TokenType"`
	PublicKey       string  `bson:"PublicKey"`
	PrivateKey      string  `bson:"PrivateKey"`
	Balance         float64 `bson:"Balance"`
	LastAddrBalance float64 `bson:"LastAddrBalance"`
}

// SetEncryptPrivateKey encrypts the given privateKey string and stores it in w.PrivateKey.
func (w *Wallet) SetEncryptPrivateKey(privateKey string) error {
	passphrase := string(w.UserID)

	// Convert the private key to bytes
	keyBytes := []byte(privateKey)

	// Create AES cipher block
	block, err := aes.NewCipher(createAESKey(passphrase))
	if err != nil {
		return fmt.Errorf("failed to create cipher: %v", err)
	}

	// Generate a random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return fmt.Errorf("failed to generate IV: %v", err)
	}

	// Encrypt using CFB mode
	ciphertext := make([]byte, len(keyBytes))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext, keyBytes)

	// Prepend IV to ciphertext
	encryptedData := append(iv, ciphertext...)

	// Base64 encode
	w.PrivateKey = base64.StdEncoding.EncodeToString(encryptedData)
	return nil
}

// GetDecryptPrivateKey decrypts w.PrivateKey and returns the original private key string.
func (w *Wallet) GetDecryptPrivateKey() (string, error) {
	passphrase := string(w.UserID)

	// Decode base64
	encryptedData, err := base64.StdEncoding.DecodeString(w.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	if len(encryptedData) < aes.BlockSize {
		return "", fmt.Errorf("encrypted data is too short")
	}

	// Extract IV and ciphertext
	iv := encryptedData[:aes.BlockSize]
	ciphertext := encryptedData[aes.BlockSize:]

	// Create AES cipher block
	block, err := aes.NewCipher(createAESKey(passphrase))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Decrypt using CFB mode
	plainText := make([]byte, len(ciphertext))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plainText, ciphertext)

	// Convert bytes to string
	return string(plainText), nil
}

// Example createAESKey function (should return a 32-byte key)
func createAESKey(passphrase string) []byte {
	key := make([]byte, 32)
	copy(key, []byte(passphrase))
	return key
}

// InsertWallet inserts a new wallet into the database
func (w *Wallet) InsertWallet() error {
	collection := GetCollection("crypto-exchange", "wallets")
	if collection == nil {
		fmt.Println("collection Faile")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, w)
	return err
}

// GetWalletsByUserID retrieves all wallets associated with a user
func GetWalletsByUserID(userID int64) ([]Wallet, error) {
	collection := GetCollection("crypto-exchange", "wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"UserID": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var wallets []Wallet
	for cursor.Next(ctx) {
		var wallet Wallet
		if err := cursor.Decode(&wallet); err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}
	return wallets, nil
}
