package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"log"
	"os"

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
	fileKey, err := os.Open(keyFile)
	if err != nil {
		c.IsEnable = false
		log.Println("Can't open key file")
		return errors.New("can't open key file")
	}
	defer fileKey.Close()
	var keyBuf []byte
	keyBuf, err = ioutil.ReadFile(keyFile)
	//log.Println(keyBuf)
	if err != nil || len(keyBuf) == 0 {
		c.IsEnable = false
		log.Println("Can't read key file or no key")
		return errors.New("can't read key file or no key")
	}
	c.key = keyBuf
	log.Println("Crypto is enabled!")
	c.IsEnable = true
	return nil
}

func (c CryptoService) IsServiceEnable() bool {
	//some data to commit
	return c.IsEnable
}

func (c CryptoService) Hash(m *models.Metrics) (int, error) {
	hasher := hmac.New(sha256.New, c.key)
	src := m.String()
	//log.Println(string(src))
	nBytes, err := hasher.Write([]byte(src))
	if err != nil {
		log.Println("Hashing error!")
		return 0, errors.New("hashing error")
	}
	//log.Printf("Signed %d bytes", nBytes)
	m.Hash = hex.EncodeToString(hasher.Sum(nil))
	//log.Println(m)
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
