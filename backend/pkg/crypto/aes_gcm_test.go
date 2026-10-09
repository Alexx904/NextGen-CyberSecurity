package crypto_test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/google/uuid"

	"nextgen-cybersec/pkg/crypto"
)

func TestNewGCMCipher_KeyLengthValidation(t *testing.T) {
	tests := []struct {
		name      string
		key       []byte
		expectErr bool
	}{
		{"nil key", nil, true},
		{"empty key", []byte{}, true},
		{"16 bytes (AES-128)", make([]byte, 16), true},
		{"24 bytes (AES-192)", make([]byte, 24), true},
		{"31 bytes", make([]byte, 31), true},
		{"32 bytes (AES-256)", make([]byte, 32), false},
		{"33 bytes", make([]byte, 33), true},
		{"64 bytes", make([]byte, 64), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := crypto.NewGCMCipher(tc.key)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("atteso errore per chiave di lunghezza %d, ottenuto nil", len(tc.key))
				}
				if !errors.Is(err, crypto.ErrInvalidKeySize) {
					t.Errorf("atteso ErrInvalidKeySize, ottenuto: %v", err)
				}
				if c != nil {
					t.Errorf("l'istanza GCMCipher deve essere nil in caso di errore")
				}
			} else {
				if err != nil {
					t.Fatalf("errore inatteso per chiave valida a 32 byte: %v", err)
				}
				if c == nil {
					t.Fatal("l'istanza GCMCipher non deve essere nil")
				}
			}
		})
	}
}

func TestGCMCipher_EncryptDecrypt_RoundTrip(t *testing.T) {
	key := make([]byte, crypto.KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generazione chiave fallita: %v", err)
	}

	cipher, err := crypto.NewGCMCipher(key)
	if err != nil {
		t.Fatalf("creazione cipher fallita: %v", err)
	}

	testCases := []struct {
		name      string
		plaintext string
		aad       string
	}{
		{"Testo semplice senza AAD", "MySecretPassword123!", ""},
		{"Testo semplice con AAD tenant", "SuperSecretSSHKeyPassphrase", "tenant-12345"},
		{"Stringa vuota", "", "tenant-empty"},
		{"Chiave privata SSH lunga simulata", "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAAGAAAAMwAAAAtzc2gtZW\n-----END OPENSSH PRIVATE KEY-----", "tenant-rsa-aad"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var aadBytes []byte
			if tc.aad != "" {
				aadBytes = []byte(tc.aad)
			}

			ciphertextB64, err := cipher.Encrypt([]byte(tc.plaintext), aadBytes)
			if err != nil {
				t.Fatalf("cifratura fallita: %v", err)
			}

			if ciphertextB64 == "" {
				t.Fatal("il ciphertext Base64 generato non deve essere vuoto")
			}

			decrypted, err := cipher.Decrypt(ciphertextB64, aadBytes)
			if err != nil {
				t.Fatalf("decifratura fallita: %v", err)
			}

			if string(decrypted) != tc.plaintext {
				t.Errorf("plaintext decifrato non corrisponde: atteso %q, ottenuto %q", tc.plaintext, string(decrypted))
			}
		})
	}
}

func TestGCMCipher_AADAuthenticationMismatch(t *testing.T) {
	key := make([]byte, crypto.KeySize)
	_, _ = rand.Read(key)

	cipher, err := crypto.NewGCMCipher(key)
	if err != nil {
		t.Fatalf("creazione cipher fallita: %v", err)
	}

	plaintext := []byte("top_secret_production_password")
	tenantA := []byte("tenant-alpha-uuid")
	tenantB := []byte("tenant-beta-uuid")

	ciphertextB64, err := cipher.Encrypt(plaintext, tenantA)
	if err != nil {
		t.Fatalf("cifratura fallita: %v", err)
	}

	// Tentativo di decifrare con tenant diverso come AAD
	_, err = cipher.Decrypt(ciphertextB64, tenantB)
	if err == nil {
		t.Fatal("la decifratura con AAD errato doveva fallire")
	}
	if !errors.Is(err, crypto.ErrDecryptionFailed) {
		t.Errorf("atteso ErrDecryptionFailed, ottenuto: %v", err)
	}

	// Tentativo di decifrare senza AAD
	_, err = cipher.Decrypt(ciphertextB64, nil)
	if err == nil {
		t.Fatal("la decifratura priva di AAD originale doveva fallire")
	}
	if !errors.Is(err, crypto.ErrDecryptionFailed) {
		t.Errorf("atteso ErrDecryptionFailed, ottenuto: %v", err)
	}
}

func TestGCMCipher_TamperDetection(t *testing.T) {
	key := make([]byte, crypto.KeySize)
	_, _ = rand.Read(key)

	cipher, _ := crypto.NewGCMCipher(key)
	ciphertextB64, _ := cipher.Encrypt([]byte("critical-system-secret"), []byte("aad-data"))

	rawBytes, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		t.Fatalf("decodifica Base64 fallita: %v", err)
	}

	// Test manomissione: modifichiamo ogni possibile posizione del ciphertext
	positionsToCorrupt := []int{
		0,                  // Nel nonce
		crypto.NonceSize,   // All'inizio del ciphertext
		len(rawBytes) - 1,  // Nell'authentication tag
	}

	for _, pos := range positionsToCorrupt {
		corrupted := make([]byte, len(rawBytes))
		copy(corrupted, rawBytes)
		corrupted[pos] ^= 0xFF // Flip bits

		corruptedB64 := base64.StdEncoding.EncodeToString(corrupted)
		_, err := cipher.Decrypt(corruptedB64, []byte("aad-data"))
		if err == nil {
			t.Errorf("la decifratura di payload manomesso alla posizione %d non doveva avere successo", pos)
		}
		if !errors.Is(err, crypto.ErrDecryptionFailed) {
			t.Errorf("atteso ErrDecryptionFailed per manomissione a %d, ottenuto: %v", pos, err)
		}
	}
}

func TestGCMCipher_MalformedCiphertext(t *testing.T) {
	key := make([]byte, crypto.KeySize)
	_, _ = rand.Read(key)

	cipher, _ := crypto.NewGCMCipher(key)

	// 1. Stringa non valida Base64
	_, err := cipher.Decrypt("%%%invalid-base64&&&", nil)
	if !errors.Is(err, crypto.ErrInvalidCiphertext) {
		t.Errorf("atteso ErrInvalidCiphertext per base64 malformato, ottenuto: %v", err)
	}

	// 2. Payload troppo breve (inferiore a nonce + tag overhead)
	shortData := base64.StdEncoding.EncodeToString([]byte("too-short"))
	_, err = cipher.Decrypt(shortData, nil)
	if !errors.Is(err, crypto.ErrInvalidCiphertext) {
		t.Errorf("atteso ErrInvalidCiphertext per dati troppo corti, ottenuto: %v", err)
	}
}

func TestStandaloneHelpers_EncryptDecrypt(t *testing.T) {
	key := make([]byte, crypto.KeySize)
	_, _ = rand.Read(key)

	secret := []byte("standalone_secret_payload")
	aad := []byte("scope-test")

	cipherB64, err := crypto.EncryptAES256GCM(secret, key, aad)
	if err != nil {
		t.Fatalf("EncryptAES256GCM fallito: %v", err)
	}

	decrypted, err := crypto.DecryptAES256GCM(cipherB64, key, aad)
	if err != nil {
		t.Fatalf("DecryptAES256GCM fallito: %v", err)
	}

	if !bytes.Equal(decrypted, secret) {
		t.Errorf("payload decifrato con helper non corrisponde")
	}

	// Test con chiave non valida
	_, err = crypto.EncryptAES256GCM(secret, []byte("short-key"), aad)
	if !errors.Is(err, crypto.ErrInvalidKeySize) {
		t.Errorf("atteso ErrInvalidKeySize, ottenuto: %v", err)
	}

	_, err = crypto.DecryptAES256GCM(cipherB64, []byte("short-key"), aad)
	if !errors.Is(err, crypto.ErrInvalidKeySize) {
		t.Errorf("atteso ErrInvalidKeySize, ottenuto: %v", err)
	}
}

func TestDeriveTenantKey(t *testing.T) {
	masterKey := []byte("master-system-secret-key-for-kdf-32b")
	tenant1 := uuid.New()
	tenant2 := uuid.New()

	// 1. Master key vuota
	_, err := crypto.DeriveTenantKey(nil, tenant1)
	if !errors.Is(err, crypto.ErrEmptyMasterKey) {
		t.Errorf("atteso ErrEmptyMasterKey per master key nil, ottenuto: %v", err)
	}

	_, err = crypto.DeriveTenantKey([]byte{}, tenant1)
	if !errors.Is(err, crypto.ErrEmptyMasterKey) {
		t.Errorf("atteso ErrEmptyMasterKey per master key vuota, ottenuto: %v", err)
	}

	// 2. Determinismo per lo stesso tenant
	key1A, err := crypto.DeriveTenantKey(masterKey, tenant1)
	if err != nil {
		t.Fatalf("derivazione key1A fallita: %v", err)
	}
	key1B, err := crypto.DeriveTenantKey(masterKey, tenant1)
	if err != nil {
		t.Fatalf("derivazione key1B fallita: %v", err)
	}
	if !bytes.Equal(key1A, key1B) {
		t.Fatal("la derivazione per lo stesso tenant deve essere deterministica")
	}
	if len(key1A) != crypto.KeySize {
		t.Fatalf("la chiave derivata deve essere di %d byte, ottenuto %d", crypto.KeySize, len(key1A))
	}

	// 3. Isolamento e diversità tra tenant distinti
	key2, err := crypto.DeriveTenantKey(masterKey, tenant2)
	if err != nil {
		t.Fatalf("derivazione key2 fallita: %v", err)
	}
	if bytes.Equal(key1A, key2) {
		t.Fatal("tenant differenti devono produrre chiavi crittografiche distinte")
	}

	// 4. Verifica decifratura cross-tenant: Tenant 2 non può decifrare dati di Tenant 1
	plaintext := []byte("root_password_production")
	cipherTenant1, err := crypto.EncryptAES256GCM(plaintext, key1A, []byte(tenant1.String()))
	if err != nil {
		t.Fatalf("cifratura tenant 1 fallita: %v", err)
	}

	_, err = crypto.DecryptAES256GCM(cipherTenant1, key2, []byte(tenant2.String()))
	if err == nil {
		t.Fatal("tenant 2 non deve essere in grado di decifrare dati di tenant 1")
	}
	if !errors.Is(err, crypto.ErrDecryptionFailed) {
		t.Errorf("atteso ErrDecryptionFailed, ottenuto: %v", err)
	}
}
