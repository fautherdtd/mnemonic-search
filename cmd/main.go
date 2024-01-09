package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"math/big"
	"mm-core/pkg/alerts"
	"mm-core/pkg/bipdump"
	"os"
	"strings"
)

func main() {
	outputFile, err := os.OpenFile("./resources/mnemonics-success.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		return
	}
	defer outputFile.Close()

	for {
		entropy, err := generateEntropy(256) // Задайте необходимую длину энтропии (в битах)
		if err != nil {
			fmt.Println("Ошибка генерации энтропии:", err)
			return
		}
		checksum := calculateChecksum(entropy)
		combinedEntropy := append(entropy, checksum...)
		groups := splitToGroups(combinedEntropy)
		words, err := lookupWords(groups)
		if err != nil {
			fmt.Println("Ошибка при поиске слов в словаре:", err)
			return
		}

		mnemonic := strings.Join(words, " ")
		result, address := bipdump.SearchAddress(mnemonic, 12)
		if result {
			outputFile.WriteString("\n")
			outputFile.WriteString(mnemonic + "   (" + address + ")")
			alerts.TelegramSend(mnemonic)
		}

	}
}
func generateEntropy(length int) ([]byte, error) {
	bytes := make([]byte, length/8)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func calculateChecksum(data []byte) []byte {
	hash := sha256.Sum256(data)
	entropyBits := len(data) * 8
	checksumBits := entropyBits / 32
	checksumBytes := checksumBits / 8
	return hash[:checksumBytes]
}

func splitToGroups(entropy []byte) []uint16 {
	groupSize := uint(11)
	groups := make([]uint16, len(entropy)*4/int(groupSize))
	bits := new(big.Int).SetBytes(entropy)

	for i := range groups {
		groups[i] = uint16(bits.Uint64() & ((1 << uint(groupSize)) - 1))
		bits.Rsh(bits, uint(groupSize))
	}

	return groups
}

func lookupWords(groups []uint16) ([]string, error) {
	godotenv.Load()
	bip := os.Getenv("BIP_DICT")
	dictionary, err := loadDictionary("resources/" + bip + ".txt")
	if err != nil {
		return nil, err
	}

	words := make([]string, len(groups))
	for i, group := range groups {
		if int(group) >= len(dictionary) {
			return nil, fmt.Errorf("Недопустимый индекс группы: %d", group)
		}
		words[i] = dictionary[group]
	}

	return words, nil
}

func loadDictionary(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var dictionary []string
	for scanner.Scan() {
		dictionary = append(dictionary, scanner.Text())
	}

	return dictionary, scanner.Err()
}
