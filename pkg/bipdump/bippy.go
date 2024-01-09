package bipdump

import (
	"fmt"
	"log"
	"mm-core/pkg/blockchain"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"

	crypto "github.com/ethereum/go-ethereum/crypto"
)

type Purpose = uint32

const (
	PurposeBIP44 Purpose = 0x8000002C // 44' BIP44
	PurposeBIP49 Purpose = 0x80000031 // 49' BIP49
	PurposeBIP84 Purpose = 0x80000054 // 84' BIP84
)

type CoinType = uint32

const (
	CoinTypeBTC CoinType = 0x80000000
	CoinTypeLTC CoinType = 0x80000002
	CoinTypeETH CoinType = 0x8000003c
	CoinTypeEOS CoinType = 0x800000c2
)

const (
	Apostrophe uint32 = 0x80000000 // 0'
)

type Key struct {
	path     string
	bip32Key *bip32.Key
}

func (k *Key) Encode(compress bool) (wif, address, segwitBech32, segwitNested string, privateKey string, err error) {
	prvKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), k.bip32Key.Key)
	return GenerateFromBytes(prvKey, compress)
}

func (k *Key) EncodeEth() (address string, publicKey string, privateKey string) {
	prvKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), k.bip32Key.Key)
	pub := pubKey.ToECDSA()
	address = crypto.PubkeyToAddress(*pub).Hex()
	publicKey = string(crypto.FromECDSAPub(pub))
	privateKey = fmt.Sprintf("%x", prvKey.D)
	return address, publicKey, privateKey
}

func (k *Key) GetPath() string {
	return k.path
}

type KeyManager struct {
	mnemonic   string
	passphrase string
	keys       map[string]*bip32.Key
	mux        sync.Mutex
}

func NewKeyManager(bitSize int, passphrase string) (*KeyManager, error) {
	entropy, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	km := &KeyManager{
		mnemonic:   mnemonic,
		passphrase: passphrase,
		keys:       make(map[string]*bip32.Key, 0),
	}
	return km, nil
}

func NewKeyManagerWithMnemonic(bitSize int, passphrase string, keyphrase string) (*KeyManager, error) {

	entropy, err := bip39.EntropyFromMnemonic(keyphrase)
	if err != nil {
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	km := &KeyManager{
		mnemonic:   mnemonic,
		passphrase: passphrase,
		keys:       make(map[string]*bip32.Key, 0),
	}
	return km, nil
}

func (km *KeyManager) GetMnemonic() string {
	return km.mnemonic
}

func (km *KeyManager) GetPassphrase() string {
	return km.passphrase
}

func (km *KeyManager) GetSeed() []byte {
	return bip39.NewSeed(km.GetMnemonic(), km.GetPassphrase())
}

func (km *KeyManager) getKey(path string) (*bip32.Key, bool) {
	km.mux.Lock()
	defer km.mux.Unlock()

	key, ok := km.keys[path]
	return key, ok
}

func (km *KeyManager) setKey(path string, key *bip32.Key) {
	km.mux.Lock()
	defer km.mux.Unlock()

	km.keys[path] = key
}

func (km *KeyManager) GetMasterKey() (*bip32.Key, error) {
	path := "m"

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	key, err := bip32.NewMasterKey(km.GetSeed())
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetPurposeKey(purpose uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'`, purpose-Apostrophe)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetMasterKey()
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(purpose)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetCoinTypeKey(purpose, coinType uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'`, purpose-Apostrophe, coinType-Apostrophe)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetPurposeKey(purpose)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(coinType)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetAccountKey(purpose, coinType, account uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'`, purpose-Apostrophe, coinType-Apostrophe, account)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetCoinTypeKey(purpose, coinType)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(account + Apostrophe)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetChangeKey(purpose, coinType, account, change uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d`, purpose-Apostrophe, coinType-Apostrophe, account, change)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetAccountKey(purpose, coinType, account)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(change)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetShortAccountKey(purpose uint32, coinType uint32, account uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d`, purpose-Apostrophe, coinType-Apostrophe, account)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	key, err := km.GetAccountKey(purpose, coinType, account)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetShortKey(purpose uint32, coinType uint32, account uint32, index uint32) (*Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d`, purpose-Apostrophe, coinType-Apostrophe, account, index)

	key, ok := km.getKey(path)
	if ok {
		return &Key{path: path, bip32Key: key}, nil
	}

	parent, err := km.GetShortAccountKey(purpose, coinType, account)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(index)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return &Key{path: path, bip32Key: key}, nil
}

func (km *KeyManager) GetKey(purpose, coinType, account, change, index uint32) (*Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d/%d`, purpose-Apostrophe, coinType-Apostrophe, account, change, index)

	key, ok := km.getKey(path)
	if ok {
		return &Key{path: path, bip32Key: key}, nil
	}

	parent, err := km.GetChangeKey(purpose, coinType, account, change)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(index)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return &Key{path: path, bip32Key: key}, nil
}

func Generate(compress bool) (wif, address, segwitBech32, segwitNested string, privateKey string, err error) {
	prvKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", "", "", "", err
	}
	return GenerateFromBytes(prvKey, compress)
}

func GenerateFromBytes(prvKey *btcec.PrivateKey, compress bool) (wif, address, segwitBech32, segwitNested string, privateKey string, err error) {
	// generate the wif(wallet import format) string
	btcwif, err := btcutil.NewWIF(prvKey, &chaincfg.MainNetParams, compress)
	if err != nil {
		return "", "", "", "", "", err
	}
	wif = btcwif.String()

	// generate a normal p2pkh address
	serializedPubKey := btcwif.SerializePubKey()
	addressPubKey, err := btcutil.NewAddressPubKey(serializedPubKey, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", "", err
	}
	address = addressPubKey.EncodeAddress()

	// generate a normal p2wkh address from the pubkey hash
	witnessProg := btcutil.Hash160(serializedPubKey)
	addressWitnessPubKeyHash, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", "", err
	}
	segwitBech32 = addressWitnessPubKeyHash.EncodeAddress()
	serializedScript, err := txscript.PayToAddrScript(addressWitnessPubKeyHash)
	if err != nil {
		return "", "", "", "", "", err
	}
	addressScriptHash, err := btcutil.NewAddressScriptHash(serializedScript, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", "", err
	}
	segwitNested = addressScriptHash.EncodeAddress()

	privateKey = fmt.Sprintf("%x", prvKey.D)

	return wif, address, segwitBech32, segwitNested, privateKey, nil
}

func SearchAddress(mnemonic string, size int) (bool, string) {
	compress := true // generate a compressed public key
	bip39 := true
	pass := ""
	number := 10
	short := true
	// mnemonic = "zero audit ceiling gain sort awkward ritual glare silver fitness tennis lucky"

	var bitsize int
	if size == 24 {
		bitsize = 256
	} else if size == 12 {
		bitsize = 128
	} else {
		log.Fatal("Invalid size, value must be 12 or 24")
		return false, ""
	}

	if !bip39 {
		fmt.Printf("\n%-34s %-34s %-52s %-42s %s\n", "Bitcoin Address", "Private Key", "WIF(Wallet Import Format)", "SegWit(bech32)", "SegWit(nested)")
		fmt.Println(strings.Repeat("-", 165))

		for i := 0; i < number; i++ {
			wif, address, segwitBech32, segwitNested, pk, err := Generate(compress)
			if err != nil {
				// log.Fatal(err)
			}
			fmt.Printf("%-34s %s %s %s %s\n", address, pk, wif, segwitBech32, segwitNested)
		}
		fmt.Println()
		return false, ""
	}

	var err error
	var km *KeyManager
	if mnemonic != "" {
		km, err = NewKeyManagerWithMnemonic(bitsize, pass, mnemonic)
	} else {
		km, err = NewKeyManager(bitsize, pass)
	}
	if err != nil {
		return false, ""
	}

	passphrase := km.GetPassphrase()
	if passphrase == "" {
		passphrase = "<none>"
	}

	for i := 0; i < number; i++ {
		var key *Key
		if short {
			key, err = km.GetShortKey(PurposeBIP44, CoinTypeBTC, 0, uint32(i))
		} else {
			key, err = km.GetKey(PurposeBIP44, CoinTypeBTC, 0, 0, uint32(i))
		}
		if err != nil {
			// log.Fatal(err)
		}
		_, btcAddress, _, _, _, err := key.Encode(compress)
		if err != nil {
			// log.Fatal(err)
		}

		// ПРОВЕРКА АДРЕСОВ
		result := blockchain.GetBalance(btcAddress, "BTC")
		if result > 0 {
			return true, btcAddress
		}
	}

	for i := 0; i < number; i++ {
		var key *Key
		if short {
			key, err = km.GetShortKey(PurposeBIP84, CoinTypeBTC, 0, uint32(i))
		} else {
			key, err = km.GetKey(PurposeBIP84, CoinTypeBTC, 0, 0, uint32(i))
		}
		if err != nil {
			// log.Fatal(err)
		}
		_, _, segwitBech32, _, _, err := key.Encode(compress)
		if err != nil {
			// log.Fatal(err)
		}

		// ПРОВЕРКА АДРЕСОВ
		result := blockchain.GetBalance(segwitBech32, "BTC")
		if result > 0 {
			return true, segwitBech32
		}
	}

	for i := 0; i < number; i++ {
		var key *Key
		if short {
			key, err = km.GetShortKey(PurposeBIP44, CoinTypeETH, 0, uint32(i))
		} else {
			key, err = km.GetKey(PurposeBIP44, CoinTypeETH, 0, 0, uint32(i))
		}
		if err != nil {
			log.Fatal(err)
		}
		address, _, _ := key.EncodeEth()

		result := blockchain.GetBalance(address, "ETH")
		if result > 0 {
			return true, address
		}
	}

	return false, ""
}
