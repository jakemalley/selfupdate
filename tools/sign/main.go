/*
Simple tool to generate ECDSA keys, and then signatures for a binary.

Usage:
	./sign [flags] <binary>
*/
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"
)

func run(generate bool, publicKeyPath, privateKeyPath, binaryPath string) error {
	if generate {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return fmt.Errorf("failed to generate keys: %w", err)
		}
		// Write private key
		privBytes, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			return fmt.Errorf("failed to marshall private key: %w", err)
		}
		privFile, err := os.Create(privateKeyPath)
		if err != nil {
			return fmt.Errorf("failed to create private key file: %w", err)
		}
		err = pem.Encode(privFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
		privFile.Close()
		if err != nil {
			return fmt.Errorf("failed to write to private key file: %w", err)
		}
		// Write public key
		pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to marshall public key: %w", err)
		}
		pubFile, err := os.Create(publicKeyPath)
		if err != nil {
			return fmt.Errorf("failed to create public key file: %w", err)
		}
		err = pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
		pubFile.Close()
		if err != nil {
			return fmt.Errorf("failed to write to public key file: %w", err)
		}
		fmt.Printf("Successfully generated public/private key pair '%s'/'%s'\n", publicKeyPath, privateKeyPath)
	}

	// Read the private key back in again
	privPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		fmt.Errorf("failed to read private key: %w", err)

	}
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block from private key: %w", err)
	}
	privKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Open binary, and compute checksum and then generate the signature
	file, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to open binary: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return fmt.Errorf("failed to hash binary: %w", err)
	}
	checksum := hash.Sum(nil)
	fmt.Printf("SHA256 checksum: %s\n", hex.EncodeToString(checksum))

	signature, err := ecdsa.SignASN1(rand.Reader, privKey, checksum)
	if err != nil {
		return fmt.Errorf("failed to sign binary: %w", err)
	}
	fmt.Printf("ECDSA Signature: %s\n", hex.EncodeToString(signature))

	return nil
}

func main() {
	var generate bool
	var publicKeyPath, privateKeyPath string
	flag.BoolVar(&generate, "generate", false, "Generate a new ECDSA key pair")
	flag.StringVar(&publicKeyPath, "public-key", "public.pem", "Path to save public key")
	flag.StringVar(&privateKeyPath, "private-key", "private.pem", "Path to save private key")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Printf("usage: %s [flags] <binary>\n", os.Args[0])
		os.Exit(2)
	}
	if err := run(generate, publicKeyPath, privateKeyPath, flag.Arg(0)); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
