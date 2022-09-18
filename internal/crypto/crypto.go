package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"

	"github.com/MaximkaSha/log_tools/internal/models"
)

type CryptoService struct {
	keyFile  string
	key      []byte
	IsEnable bool
}

// Read Key from file to var
func NewCryptoService() CryptoService {
	return CryptoService{
		IsEnable: false,
	}
}

func (c *CryptoService) InitCryptoService(keyFile string) error {
	/*	fileKey, err := os.Open(keyFile)
		if err != nil {
			c.IsEnable = false
			log.Printf("Can't open key file %s", keyFile)
			return errors.New("can't open key file")
		}
		defer fileKey.Close() */
	keyBuf := []byte(keyFile)
	if len(keyBuf) == 0 {
		c.IsEnable = false
		log.Println("no key")
		return errors.New("no key")
	}
	c.key = keyBuf
	log.Println("Crypto is enabled!")
	c.IsEnable = true
	return nil
}

func (c CryptoService) IsServiceEnable() bool {
	return c.IsEnable
}

func (c CryptoService) Hash(m *models.Metrics) (int, error) {
	hasher := hmac.New(sha256.New, c.key)
	src := m.StringData()
	nBytes, err := hasher.Write([]byte(src))
	if err != nil {
		log.Println("Hashing error!")
		return 0, errors.New("hashing error")
	}
	m.Hash = hex.EncodeToString(hasher.Sum(nil))
	return nBytes, nil
}

func (c CryptoService) CheckHash(m models.Metrics) bool {
	hasher := hmac.New(sha256.New, c.key)
	src := m.StringData()
	_, err := hasher.Write([]byte(src))
	if err != nil {
		log.Println("Hashing error!")
		return false
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	if string(hash) != m.Hash {
		log.Println("Sign check false")
		return false
	}
	return true
}
