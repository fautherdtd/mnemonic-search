# Hierarchical Deterministic (HD) addresses 

Generates mnemonic phrases from a rotated dictionary. The received addresses are dumped and searched for balance on the Blockchain

Select a dictionary to search for and add the name to .env
* Standard dictionary - <code>bip-0039</code>
* Dictionary turned upside down - <code>bip-random-0039</code>

example: 
<code>BIP_CHANGE=bip-random-0039</code>

<br>Golang implementation of the BIP32/BIP39/BIP43/BIP44/SLIP44/BIP49/BIP84/BIP173  recovering keys, mnemonic seeds and Hierarchical Deterministic (HD) addresses.

<br>If the search is successful, it sends a signal to Telegram.
* [Create Telegram Bot](https://core.telegram.org/bots/tutorial)

<br>Save token and chat_id in .env
* <code>TELEGRAM_TOKEN=</code>
* <code>TELEGRAM_CHAT_ID=</code>

The successful result of the mnemonic phrase is also saved to a file with the address
* resources/mnemonics-success.txt
### go run cmd/main.go