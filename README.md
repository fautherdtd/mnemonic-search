# Hierarchical Deterministic (HD) addresses 

Generates mnemonic phrases from a rotated dictionary. The received addresses are dumped and searched for balance on the Blockchain

* Standard dictionary - <b>resources/bip-0039.txt</b>
* Dictionary turned upside down - <b>resources/bip-random-0039.txt</b>

<br>Golang implementation of the BIP32/BIP39/BIP43/BIP44/SLIP44/BIP49/BIP84/BIP173  recovering keys, mnemonic seeds and Hierarchical Deterministic (HD) addresses.

<br>If the search is successful, it sends a signal to Telegram.
* [Create Telegram Bot](https://core.telegram.org/bots/tutorial)

<br>Save token and chat_id in .env
* TELEGRAM_TOKEN
* TELEGRAM_CHAT_ID

The successful result of the mnemonic phrase is also saved to a file with the address
* resources/mnemonics-success.txt
### go run cmd/main.go