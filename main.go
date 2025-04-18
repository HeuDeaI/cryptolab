package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

const (
	keyFile       = "keys.bin"
	sBoxFile      = "table.bin"
	inputFile     = "data.bin"
	encryptedFile = "encrypted.bin"
	decryptedFile = "decrypted.bin"
)

func rotateLeft11(x uint32) uint32 {
	return (x << 11) | (x >> (32 - 11))
}

func gostF(n, k uint32, sBox [][]uint32) uint32 {
	temp := (n + k) & 0xFFFFFFFF
	sRes := uint32(0)
	for i := 0; i < 8; i++ {
		val := (temp >> (4 * i)) & 0xF
		sRes |= (sBox[i][val] << (4 * i))
	}
	return rotateLeft11(sRes)
}

func encryptBlock(n1, n2 uint32, key []uint32, sBox [][]uint32) (uint32, uint32) {
	for i := 0; i < 24; i++ {
		temp := n1
		n1 = n2 ^ gostF(n1, key[i%8], sBox)
		n2 = temp
	}
	for i := 7; i >= 0; i-- {
		temp := n1
		n1 = n2 ^ gostF(n1, key[i], sBox)
		n2 = temp
	}
	return n2, n1
}

func decryptBlock(n1, n2 uint32, key []uint32, sBox [][]uint32) (uint32, uint32) {
	for i := 0; i < 8; i++ {
		temp := n1
		n1 = n2 ^ gostF(n1, key[i], sBox)
		n2 = temp
	}
	for i := 23; i >= 0; i-- {
		temp := n1
		n1 = n2 ^ gostF(n1, key[i%8], sBox)
		n2 = temp
	}
	return n2, n1
}

func gostEncrypt(data []byte, key []uint32, sBox [][]uint32) []byte {
	var encrypted []byte
	blockSize := 8
	for i := 0; i < len(data); i += blockSize {
		block := make([]byte, blockSize)
		copy(block, data[i:min(i+blockSize, len(data))])

		n1 := binary.LittleEndian.Uint32(block[:4])
		n2 := binary.LittleEndian.Uint32(block[4:])

		r1, r2 := encryptBlock(n1, n2, key, sBox)

		r1Bytes := make([]byte, 4)
		r2Bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(r1Bytes, r1)
		binary.LittleEndian.PutUint32(r2Bytes, r2)

		encrypted = append(encrypted, r1Bytes...)
		encrypted = append(encrypted, r2Bytes...)
	}
	return encrypted
}

func gostDecrypt(data []byte, key []uint32, sBox [][]uint32) []byte {
	var decrypted []byte
	blockSize := 8
	for i := 0; i < len(data); i += blockSize {
		block := data[i:min(i+blockSize, len(data))]

		n1 := binary.LittleEndian.Uint32(block[:4])
		n2 := binary.LittleEndian.Uint32(block[4:])

		r1, r2 := decryptBlock(n1, n2, key, sBox)

		r1Bytes := make([]byte, 4)
		r2Bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(r1Bytes, r1)
		binary.LittleEndian.PutUint32(r2Bytes, r2)

		decrypted = append(decrypted, r1Bytes...)
		decrypted = append(decrypted, r2Bytes...)
	}
	return decrypted
}

func readKey(filename string) ([]uint32, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if len(data) != 32 {
		return nil, fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(data))
	}

	key := make([]uint32, 8)
	for i := 0; i < 8; i++ {
		key[i] = binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])
	}
	return key, nil
}

func readSBox(filename string) ([][]uint32, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if len(data) < 64 {
		return nil, fmt.Errorf("invalid sbox length: expected at least 64 bytes, got %d", len(data))
	}

	sBox := make([][]uint32, 8)
	for i := 0; i < 8; i++ {
		sBox[i] = make([]uint32, 16)
		for j := 0; j < 8; j++ {
			b := data[i*8+j]
			sBox[i][j*2] = uint32(b & 0x0F)
			sBox[i][j*2+1] = uint32((b >> 4) & 0x0F)
		}
	}
	return sBox, nil
}

func readInput(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	padLen := (8 - (len(data) % 8)) % 8
	for i := 0; i < padLen; i++ {
		data = append(data, byte(padLen))
	}
	return data, nil
}

func writeOutput(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

func main() {
	fmt.Println("GOST Encryption/Decryption Tool")
	fmt.Print("Choose mode:\n1 - Encrypt\n0 - Decrypt\n> ")

	var mode int
	if _, err := fmt.Scan(&mode); err != nil {
		log.Fatal("Error reading mode:", err)
	}

	key, err := readKey(keyFile)
	if err != nil {
		log.Fatal("Failed to read key:", err)
	}

	sBox, err := readSBox(sBoxFile)
	if err != nil {
		log.Fatal("Failed to read S-box:", err)
	}

	switch mode {
	case 1:
		data, err := readInput(inputFile)
		if err != nil {
			log.Fatal("Error reading input:", err)
		}

		encrypted := gostEncrypt(data, key, sBox)
		if err := writeOutput(encryptedFile, encrypted); err != nil {
			log.Fatal("Error writing encrypted data:", err)
		}
		fmt.Printf("Data encrypted and saved to %s\n", encryptedFile)

	case 0:
		data, err := readInput(encryptedFile)
		if err != nil {
			log.Fatal("Error reading encrypted data:", err)
		}

		decrypted := gostDecrypt(data, key, sBox)
		if err := writeOutput(decryptedFile, decrypted); err != nil {
			log.Fatal("Error writing decrypted data:", err)
		}
		fmt.Printf("Data decrypted and saved to %s\n", decryptedFile)

	default:
		log.Fatal("Invalid mode selected")
	}

	fmt.Println("Operation completed.")
}
