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
	Expiration       int    `json:"expiration"`
	Height           int    `json:"activeHeight"`
	Priority         int    `json:"priority"`
	OraclePrivateKey string `json:"oraclePrivateKey"`
	ChainID          string `json:"chainID"`
	PayingKeyName    string `json:"payingKeyName"`
}

type Outputs struct {
	Version    int `json:"ver"`
	Expiration int `json:"expiration"`
	Height     int `json:"activeHeight"`
	Price      int `json:"price"`
	Priority   int `json:"priority"`
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
		outs.Expiration = ins.Expiration
		outs.Height = ins.Height
		outs.Price = ins.Price
		outs.Priority = ins.Priority

		outsMarshalledBytes, _ := json.Marshal(outs)
		outsMarshalled := string(outsMarshalledBytes)
		fmt.Println("Price message:")
		fmt.Println(outsMarshalled)

		signature := ed25519.Sign(privateKey, outsMarshalledBytes)

		fmt.Printf("signature of price message: %x\n\n", signature[:])

		//fmt.Printf("echo -n '%s' | factom-cli put -e %x -c %s %s\n", outsMarshalled, signature[:], ins.ChainID, ins.PayingKeyName)
		//factom-cli might put ASCII encoded text in the signature field instead of the binary signature.  Might need to use curl API instead.
                
                fmt.Printf("curl -i -X POST -H 'Content-Type: application/json' -d '{\"ChainID\":\"%s\", \"ExtIDs\":[\"%x\"], \"Content\":\"48656C6C6F20466163746F6D21\"}' localhost:8089/v1/compose-entry-submit/zeros",  ins.ChainID, signature[:])

                
                //curl -i -X POST -H 'Content-Type: application/json' -d '{"ChainID":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "ExtIDs":["4b007494cb7d985cf18f651d7ce50d88390f83846c46ae3c01b543c73c828bd7734dbcb7aa583cfb080e1421c255b471bf8660d995e43a16a44de8306e083001"], "Content":"48656C6C6F20466163746F6D21"}' localhost:8089/v1/compose-entry-submit/zeros | python -c "import json,sys;[line for line in sys.stdin]; obj=json.loads(line); print obj['EntryCommit']['CommitEntryMsg'];"

	}

}

func createfile() {
	var example_file = `{
  "price": 88888,
  "price_comment": "This the integer number of Factoshis required to purchase 1 Entry Credit.  Calculate by ((100000/fct price)/bitcoin price) ((100000/.0025)/450)",

  "expiration": 50006,
  "expiration_comment": "This message must appear in this block, or in the previous 6 blocks (inclusive) in order to be valid.",

  "activeHeight": 50000,
  "activeHeight_comment": "This is the block height where this price should take effect if published with enough lead time.  It must be within +- 6 of the expiration.",

  "priority": 0,
  "priority_comment": "If there is a higher priority valid message published in the previous 12 blocks (or the current block) then the higher priority one is used.",

  "oraclePrivateKey": "2e66aece65eb01ed3106dbcf0f2ea1cfac0bff7161b05575799c835635167fe5",
  "oraclePrivateKey_comment": "This is the 32 byte ed25519 private key which the federated servers will respect as the data source for the price feed",

  "chainID": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "chainID_comment": "This is the chain where the price feed will live.  It is only used to construct the factom-cli command.",

  "payingKeyName": "zeros",
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
