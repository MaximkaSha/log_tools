// Package keysgenerator generates RSA 4096 key pair in PEM format for agent and server.
package main

import (
	"fmt"
	"os"

	"github.com/MaximkaSha/log_tools/internal/ciphers"
)

func main() {
	priv, pub := ciphers.GenerateKeyPair(4096)
	privPem := ciphers.ExportRsaPrivateKeyAsPemStr(priv)
	pubPem := ciphers.PublicKeyToBytes(pub)
	pubKey, err := os.Create("pub.key")
	if err != nil {
		fmt.Println(err)
		pubKey.Close()
		return
	}
	privKey, err := os.Create("priv.key")
	if err != nil {
		fmt.Println(err)
		privKey.Close()
		return
	}
	fmt.Fprintln(pubKey, pubPem)
	fmt.Fprintln(privKey, privPem)
	err = pubKey.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = privKey.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Key generated successfully")

}
