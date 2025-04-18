package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"os"
)

func createKeyFile(filename string) error {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return fmt.Errorf("error generating key: %v", err)
	}
	return os.WriteFile(filename, key, 0644)
}

func createTableFile(filename string) error {
	sbox := [][]uint8{
		{4, 10, 9, 2, 13, 8, 0, 14, 6, 11, 1, 12, 7, 15, 5, 3},
		{14, 11, 4, 12, 6, 13, 15, 10, 2, 3, 8, 1, 0, 7, 5, 9},
		{5, 8, 1, 13, 10, 3, 4, 2, 14, 15, 12, 7, 6, 0, 9, 11},
		{7, 13, 10, 1, 0, 8, 9, 15, 14, 4, 6, 12, 11, 2, 5, 3},
		{6, 12, 7, 1, 5, 15, 13, 8, 4, 10, 9, 14, 0, 3, 11, 2},
		{4, 11, 10, 0, 7, 2, 1, 13, 3, 6, 8, 5, 9, 12, 15, 14},
		{13, 11, 4, 1, 3, 15, 5, 9, 0, 10, 14, 7, 6, 8, 2, 12},
		{1, 15, 13, 0, 5, 7, 10, 4, 9, 2, 3, 14, 6, 11, 8, 12},
	}
	var buf bytes.Buffer
	for _, row := range sbox {
		for i := 0; i < len(row); i += 2 {
			b := (row[i+1] << 4) | row[i]
			buf.WriteByte(b)
		}
	}
	return os.WriteFile(filename, buf.Bytes(), 0644)
}

func createDataFile(filename string) error {
	data := []byte("Hello, GOST!")
	return os.WriteFile(filename, data, 0644)
}

func createEmptyFile(filename string) error {
	return os.WriteFile(filename, []byte{}, 0644)
}

func main() {
	keyFile := "keys.bin"
	sBoxFile := "table.bin"
	inputFile := "data.bin"
	encryptedFile := "encrypted.bin"
	decryptedFile := "decrypted.bin"

	if err := createKeyFile(keyFile); err != nil {
		log.Fatalf("Error creating file %s: %v", keyFile, err)
	}
	log.Printf("%s successfully created.", keyFile)

	if err := createTableFile(sBoxFile); err != nil {
		log.Fatalf("Error creating file %s: %v", sBoxFile, err)
	}
	log.Printf("%s successfully created.", sBoxFile)

	if err := createDataFile(inputFile); err != nil {
		log.Fatalf("Error creating file %s: %v", inputFile, err)
	}
	log.Printf("%s successfully created.", inputFile)

	if err := createEmptyFile(encryptedFile); err != nil {
		log.Fatalf("Error creating file %s: %v", encryptedFile, err)
	}
	log.Printf("%s successfully created.", encryptedFile)

	if err := createEmptyFile(decryptedFile); err != nil {
		log.Fatalf("Error creating file %s: %v", decryptedFile, err)
	}
	log.Printf("%s successfully created.", decryptedFile)
}
