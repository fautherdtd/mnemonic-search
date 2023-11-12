package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type BalanceBTC struct {
	Confirmed int64 `json:"confirmed"`
}
type BalanceETH struct {
	Balance int64 `json:"balance"`
}

func GetBalance(address string) int64 {
	url := "https://api.blockchain.info/haskoin-store/btc/address/" + address + "/balance"

	// Отправляем GET-запрос на API
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	// Проверяем статус-код
	if response.StatusCode != http.StatusOK {
		fmt.Println("Неверный статус-код: %d", response.StatusCode)
	}

	// Чтение и декодирование JSON-ответа
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	var balanceBTC BalanceBTC
	err = json.Unmarshal(body, &balanceBTC)
	if err != nil {
		fmt.Println("Неверный BTC адрес: ", address)
	}
	fmt.Println("Адрес:", address, " Баланс: ", balanceBTC.Confirmed)
	return balanceBTC.Confirmed
}
