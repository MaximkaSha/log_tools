// Package keysgenerator generates RSA 4096 key pair in PEM format for agent and server.
package main

import (
	"fmt"
	"os"

	"github.com/MaximkaSha/log_tools/internal/ciphers"
)

func main() {
	priv, pub := ciphers.GenerateKeyPair(4096)
	priv_pem := ciphers.ExportRsaPrivateKeyAsPemStr(priv)
	pub_pem, _ := ciphers.ExportRsaPublicKeyAsPemStr(pub)
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
	fmt.Fprintln(pubKey, pub_pem)
	fmt.Fprintln(privKey, priv_pem)
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
	/*
		privNew, err := ciphers.ReadPrivateKeyFromFile("priv.key")
		if err != nil {
			log.Panicln(err)
		}
		log.Println(ciphers.ExportRsaPrivateKeyAsPemStr(privNew))
		pubNew, err := ciphers.ReadPublicKeyFromFile("pub.key")
		if err != nil {
			log.Panic(err)
		}
		log.Println(ciphers.ExportRsaPublicKeyAsPemStr(pubNew)) */
}
