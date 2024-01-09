package blockchain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type BalanceBTC struct {
	Confirmed int64 `json:"confirmed"`
}

type BalanceETH struct {
	Balance string `json:"balance"`
}

func GetBalance(address string, coin string) int64 {
	var url string
	var balanceBTC BalanceBTC
	var balanceETH BalanceETH

	defer func() {
		balanceBTC = BalanceBTC{}
		balanceETH = BalanceETH{}
	}()

	if coin == "BTC" {
		url = "https://api.blockchain.info/haskoin-store/btc/address/" + address + "/balance"
	}
	if coin == "ETH" {
		url = "https://api.blockchain.info/v2/eth/data/account/" + address + "/wallet"
	}

	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Неверный статус-код:", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	if coin == "BTC" {
		err = json.Unmarshal(body, &balanceBTC)
		if err != nil {
			fmt.Println("Неверный BTC адрес:", address)
		}
		fmt.Println("Адрес:", address, " Баланс: ", balanceBTC.Confirmed)
		return balanceBTC.Confirmed
	}

	if coin == "ETH" {
		err = json.Unmarshal(body, &balanceETH)
		if err != nil {
			fmt.Println("Неверный ETH адрес:", err)
		}
		fmt.Println("Адрес:", address, " Баланс: ", balanceETH.Balance)
		i, _ := strconv.ParseInt(balanceETH.Balance, 10, 64)
		return i
	}
	return 0
}
