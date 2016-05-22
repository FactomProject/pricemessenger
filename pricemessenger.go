package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/ed25519"
	"io"
	"io/ioutil"
	"os"
)

type Inputs struct {
	Price            int    `json:"price"`
	Height           int    `json:"height"`
	Priority         int    `json:"priority"`
	OraclePrivateKey string `json:"oraclePrivateKey"`
	ChainID          string `json:"chainID"`
	PayingKeyName    string `json:"payingKeyName"`
}

type Outputs struct {
	Version  int `json:"ver"`
	Price    int `json:"price"`
	Height   int `json:"height"`
	Priority int `json:"priority"`
}

func main() {

	//Check if file exists

	if _, err := os.Stat("setprices.txt"); os.IsNotExist(err) {
		//make the example file
		fmt.Println("Could not find the setprices.txt file.  Creating the default one now.")
		createfile()
	} else { //if the file exists
		fi, err := os.Open("setprices.txt")
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := fi.Close(); err != nil {
				panic(err)
			}
		}()

		jsonData, err := ioutil.ReadAll(fi)

		if err != nil {
			fmt.Println("Error reading JSON data:", err)
			return
		}

		var ins Inputs

		if err := json.Unmarshal(jsonData, &ins); err != nil {
			fmt.Println("Error parsing JSON data:", err)
			return
		}
		//fmt.Println(ins)

		impliedFctPrice := 100000 / float64(ins.Price)
		fmt.Printf("implied factoid price: $%.2f\n", impliedFctPrice)

		shortPrivate, err := hex.DecodeString(ins.OraclePrivateKey)
		if err != nil {
			fmt.Println("Error reading private key from file:", err)
			return
		}

		publicKey := new([32]byte)
		privateKey := new([64]byte)

		copy(privateKey[:32], shortPrivate[32:])
		publicKey = ed25519.GetPublicKey(privateKey)

		fmt.Printf("using oracle public key: %x\n", publicKey[:])

		//make the json datastructure for publishing
		var outs Outputs

		outs.Version = 0
		outs.Price = ins.Price
		outs.Height = ins.Height
		outs.Priority = ins.Priority

		outsMarshalledBytes, _ := json.Marshal(outs)
		outsMarshalled := string(outsMarshalledBytes)
		fmt.Println("Price message:")
		fmt.Println(outsMarshalled)

		signature := ed25519.Sign(privateKey, outsMarshalledBytes)

		fmt.Printf("signature of price message: %x\n\n", signature[:])

		fmt.Printf("echo -n '%s' | factom-cli put -e %x -c %s %s\n", outsMarshalled, signature[:], ins.ChainID, ins.PayingKeyName)
		//factom-cli might put ASCII encoded text in the signature field instead of the binary signature.  Might need to use curl API instead.

	}

}

func createfile() {
	var example_file = `{
  "price": 88888,
  "price_comment": "This the integer number of Factoshis required to purchase 1 Entry Credit.  calculate by ((100000/fct price)/bitcoin price) ((100000/.0025)/450)",

  "height": 50000,
  "height_comment": "This is the block height where this price should take effect if published in the right time frame.",

  "priority": 0,
  "priority_comment": "If there are multiple valid price settings at a particular height, the higher priority will take precedence. Maximum is 1000",

  "oraclePrivateKey": "2e66aece65eb01ed3106dbcf0f2ea1cfac0bff7161b05575799c835635167fe5",
  "oraclePrivateKey_comment": "This is the 32 byte ed25519 private key which the federated servers will respect as the data source for the price feed",

  "chainID": "782048258a1660961b30261e1fb568e372f6aa6eeba13043104e1e6bbcb73357",
  "chainID_comment": "This is the chain where the price feed will live.  It is only used to construct the factom-cli command.",

  "payingKeyName": "factoidprice",
  "payingKeyName_comment": "This is purely for ease of use. the example factom-cli command will use this address to pay for the entry"
}
`

	f, err := os.Create("setprices.txt")
	if err != nil {
		fmt.Println(err)
	}
	n, err := io.WriteString(f, example_file)
	if err != nil {
		fmt.Println(n, err)
	}
	f.Close()

}
