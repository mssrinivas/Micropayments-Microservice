package main
import (
	"fmt"
	"encoding/json"
	"log"
	"time"
	"net/http"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/gorilla/mux"
)

type Balance struct {
    bal uint64 
}

// Generate Keys
func GenKeys(w http.ResponseWriter, r *http.Request) {
	secret, mnemonic := hedera.GenerateSecretKey()
	fmt.Printf("secret   = %v\n", secret)
	fmt.Printf("mnemonic = %v\n", mnemonic)
	public := secret.Public()
	fmt.Printf("public   = %v\n", public)

}

// Function to fetch balance
func GetMyBalance(w http.ResponseWriter, r *http.Request) {
	accountID := hedera.AccountID{Account: <SENDER_ACC_ID>}

	client, err := hedera.Dial("testnet.hedera.com:xxxx3")
	if err != nil {
		panic(err)
	}

	client.SetNode(hedera.AccountID{Account: <NODE_NUMBER>})
	client.SetOperator(accountID, func() hedera.SecretKey {
		operatorSecret, err := hedera.SecretKeyFromString("<SECRET KEY>")
		if err != nil {
			panic(err)
		}

		return operatorSecret
	})

	defer client.Close()

	balance, err := client.Account(accountID).Balance().Get()
	if err != nil {
		panic(err)
	}

	respondWithJson(w, http.StatusOK, map[string]uint64{"Balance": balance/100000000.0})


}
// Function to make a transaction
func TransferTokens(w http.ResponseWriter, r *http.Request) {
	// Read and decode the operator secret key
	operatorAccountID := hedera.AccountID{Account: 1001}
	operatorSecret, err := hedera.SecretKeyFromString("<SECRET_KEY>")
	if err != nil {
		panic(err)
	}

	// Read and decode target account
	targetAccountID, err := hedera.AccountIDFromString("0:0:1004") // eg: 0:0:1000
	if err != nil {
		panic(err)
	}

	client, err := hedera.Dial("testnet.hedera.com:****3")
	if err != nil {
		panic(err)
	}

	client.SetNode(hedera.AccountID{Account: <NODE_NUMBER>})
	client.SetOperator(operatorAccountID, func() hedera.SecretKey {
		return operatorSecret
	})

	defer client.Close()

	balance, err := client.Account(targetAccountID).Balance().Get()
	if err != nil {
		panic(err)
	}

	nodeAccountID := hedera.AccountID{Account: <NODE_NUMBER>}
	response, err := client.TransferCrypto().
		Transfer(operatorAccountID, -100000000).
		Transfer(targetAccountID, 100000000).
		Operator(operatorAccountID).
		Node(nodeAccountID).
		Memo("[test] hedera-sdk-go v2").
		Sign(operatorSecret).
		Sign(operatorSecret).
		Execute()

	if err != nil {
		panic(err)
	}

	transactionID := response.ID
	fmt.Printf("transferred; transaction = %v\n", transactionID)
	time.Sleep(2 * time.Second)

	receipt, err := client.Transaction(*transactionID).Receipt().Get()
	if err != nil {
		panic(err)
	}

	if receipt.Status != hedera.StatusSuccess {
		panic(fmt.Errorf("transaction has a non-successful status: %v", receipt.Status.String()))
	}
	time.Sleep(2 * time.Second)

	balance, err = client.Account(operatorAccountID).Balance().Get()
	if err != nil {
		panic(err)
	}
	respondWithJson(w, http.StatusOK, map[string]uint64{"Balance": balance/100000000.0})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/generatekeys", GenKeys).Methods("GET")
	r.HandleFunc("/getbalance", GetMyBalance).Methods("GET")
	r.HandleFunc("/pay", TransferTokens).Methods("POST")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
