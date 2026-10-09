package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
)

var (
	// ErrInvalidKeySize indica che la chiave fornita non rispetta la lunghezza di 32 byte richiesta da AES-256
	ErrInvalidKeySize = errors.New("chiave crittografica non valida: attesi esattamente 32 byte (256 bit)")

	// ErrInvalidCiphertext indica che il testo cifrato è malformato, incompleto o corrotto
	ErrInvalidCiphertext = errors.New("testo cifrato non valido o corrotto")

	// ErrDecryptionFailed indica un fallimento dell'autenticazione AEAD (tag o AAD non validi)
	ErrDecryptionFailed = errors.New("decifratura fallita: integrità o autenticazione dei dati compromessa")

	// ErrEmptyMasterKey indica che la master key di sistema fornita per la derivazione è vuota
	ErrEmptyMasterKey = errors.New("master key per la derivazione non può essere vuota")
)

const (
	// KeySize definisce la dimensione in byte della chiave AES-256 (256 bit)
	KeySize = 32

	// NonceSize definisce la dimensione del nonce conforme a NIST SP 800-38D (96 bit)
	NonceSize = 12
)

// GCMCipher gestisce le operazioni di crittografia autenticata AES-256-GCM
type GCMCipher struct {
	key [KeySize]byte
}

// NewGCMCipher crea un'istanza di GCMCipher previa validazione della chiave a 256 bit
func NewGCMCipher(key []byte) (*GCMCipher, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("%w: ricevuti %d byte", ErrInvalidKeySize, len(key))
	}

	c := &GCMCipher{}
	copy(c.key[:], key)
	return c, nil
}

// Encrypt cifra il plaintext e autentica opzionalmente i dati aggiuntivi (AAD).
// Restituisce una stringa Base64 contenente il nonce anteposto al ciphertext e al tag di autenticazione.
func (c *GCMCipher) Encrypt(plaintext []byte, additionalData []byte) (string, error) {
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return "", fmt.Errorf("creazione block cipher AES fallita: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("inizializzazione Galois Counter Mode (GCM) fallita: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generazione nonce casuale fallita: %w", err)
	}

	// Seal aggiunge nonce || ciphertext + tag (16 byte)
	sealed := gcm.Seal(nonce, nonce, plaintext, additionalData)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decodifica il payload Base64, estrae il nonce e decifra il ciphertext verificando l'integrità e l'AAD.
func (c *GCMCipher) Decrypt(ciphertextB64 string, additionalData []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("%w: decodifica base64 fallita: %v", ErrInvalidCiphertext, err)
	}

	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return nil, fmt.Errorf("creazione block cipher AES fallita: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("inizializzazione Galois Counter Mode (GCM) fallita: %w", err)
	}

	nonceSize := gcm.NonceSize()
	minPayloadSize := nonceSize + gcm.Overhead()
	if len(data) < minPayloadSize {
		return nil, fmt.Errorf("%w: lunghezza dati %d inferiore al minimo %d byte", ErrInvalidCiphertext, len(data), minPayloadSize)
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptAES256GCM è una funzione helper che incapsula inizializzazione ed encryption in un'unica chiamata.
func EncryptAES256GCM(plaintext, key, additionalData []byte) (string, error) {
	cipherInstance, err := NewGCMCipher(key)
	if err != nil {
		return "", err
	}
	return cipherInstance.Encrypt(plaintext, additionalData)
}

// DecryptAES256GCM è una funzione helper che incapsula inizializzazione e decryption in un'unica chiamata.
func DecryptAES256GCM(ciphertextB64 string, key, additionalData []byte) ([]byte, error) {
	cipherInstance, err := NewGCMCipher(key)
	if err != nil {
		return nil, err
	}
	return cipherInstance.Decrypt(ciphertextB64, additionalData)
}

// DeriveTenantKey calcola una chiave a 256 bit isolata e univoca per tenant a partire dalla master key
// mediante HMAC-SHA256, conformemente a NFR-SEC-06.
func DeriveTenantKey(masterKey []byte, tenantID uuid.UUID) ([]byte, error) {
	if len(masterKey) == 0 {
		return nil, ErrEmptyMasterKey
	}

	mac := hmac.New(sha256.New, masterKey)
	contextString := fmt.Sprintf("nextgen:tenant-ssh-secret:%s", tenantID.String())
	mac.Write([]byte(contextString))
	return mac.Sum(nil), nil
}
