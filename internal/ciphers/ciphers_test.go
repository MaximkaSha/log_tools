package ciphers

import (
	"crypto/rsa"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKeyPair(t *testing.T) {
	type args struct {
		bits int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "pos1",
			args: args{bits: 4096},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pub1, priv1 := GenerateKeyPair(tt.args.bits)
			pub2, priv2 := GenerateKeyPair(tt.args.bits)
			assert.NotEqual(t, pub1, pub2, "Keys are equal")
			assert.NotEqual(t, priv1, priv2, "Keys are equal")
		})
	}
}

func TestPrivateKeyToBytes(t *testing.T) {

	tests := []struct {
		name string
	}{
		{
			name: "pos1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privKey, _ := GenerateKeyPair(4096)
			bytePriviteKey := PrivateKeyToBytes(privKey)
			newKey, err := ParseRsaPrivateKeyFromPemStr(string(bytePriviteKey))
			assert.NoErrorf(t, err, "PrivateKeyToBytes fail")
			assert.Equal(t, newKey, privKey, "PrivateKeyToBytes fail")
			_, err = ParseRsaPrivateKeyFromPemStr("not a key")
			assert.Errorf(t, err, "PublicKeyToBytes fail")
			pemStr := ExportRsaPrivateKeyAsPemStr(privKey)
			newKey, err = ParseRsaPrivateKeyFromPemStr(pemStr)
			assert.NoErrorf(t, err, "PrivateKeyToBytes fail")
			assert.Equal(t, newKey, privKey, "PrivateKeyToBytes fail")

		})
	}
}

func TestPublicKeyToBytes(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "pos1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, pubKey := GenerateKeyPair(4096)
			bytePublicKey := PublicKeyToBytes(pubKey)
			newKey, err := ParseRsaPublicKeyFromPemStr(string(bytePublicKey))
			assert.NoErrorf(t, err, "PublicKeyToBytes fail")
			assert.Equal(t, newKey, pubKey, "PublicKeyToBytes fail")
			_, err = ParseRsaPublicKeyFromPemStr("not a key")
			assert.Errorf(t, err, "PublicKeyToBytes fail")
		})
	}
}

func TestEncryptWithPublicKey(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "pos1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privKey, pubKey := GenerateKeyPair(4096)
			cipheredData := EncryptWithPublicKey([]byte(`some very secret data`), pubKey)
			plainText := DecryptWithPrivateKey(cipheredData, privKey)
			assert.Equal(t, []byte(`some very secret data`), plainText)

		})
	}
}

func TestReadPublicKeyFromFile(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "pos1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privKey, pubKey := GenerateKeyPair(4096)
			privKeyBytes := PrivateKeyToBytes(privKey)
			pubKeyBytes := PublicKeyToBytes(pubKey)
			os.WriteFile("priv.key", privKeyBytes, fs.FileMode(os.O_RDWR))
			os.WriteFile("pub.key", pubKeyBytes, fs.FileMode(os.O_RDWR))
			readedPriviteKey, err := ReadPrivateKeyFromFile("priv.key")
			assert.NoError(t, err)
			assert.Equal(t, readedPriviteKey, privKey)
			readedPublicKey, err := ReadPublicKeyFromFile("pub.key")
			assert.NoError(t, err)
			assert.Equal(t, readedPublicKey, pubKey)
			os.Remove("priv.key")
			os.Remove("pub.key")
		})
	}
}

func TestGenerateTLSCert(t *testing.T) {
	type args struct {
		key rsa.PrivateKey
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "pos1",
			args: args{
				key: rsa.PrivateKey{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GenerateTLSCert(tt.args.key)
			got2, got3 := GenerateTLSCert(tt.args.key)
			assert.NotEqual(t, got, got2, "Certs are equal")
			assert.NotEqual(t, got1, got3, "Certs are equal")
			os.Remove("cert.pem")
		})
	}
}
