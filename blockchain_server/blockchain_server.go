package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/oniwa-shuto/blockchain/block"
	"github.com/oniwa-shuto/blockchain/utils"
	"github.com/oniwa-shuto/blockchain/wallet"
)

var chash map[string]*block.BlockChain = make(map[string]*block.BlockChain)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}

func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}

func (bcs *BlockchainServer) GetBlockchain() *block.BlockChain {
	bc, ok := chash["blockchain"]
	if !ok {
		minersWallet := wallet.NewWallet()
		bc = block.NewBlockChain(minersWallet.BlockchainAddress(), bcs.Port())
		chash["blockchain"] = bc
	}
	return bc
}

func (bcs *BlockchainServer) GetChain(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()
		m, _ := bc.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		log.Printf("ERROR: Invalid HTTP Method 1")
	}
}

func (bcs *BlockchainServer) Transaction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()
		transactions := bc.TransactionPool()
		m, _ := json.Marshal(struct {
			Transaction []*block.Transaction `json:"transaction"`
			Length      int                  `json:"length"`
		}{
			Transaction: transactions,
			Length:      len(transactions),
		})
		io.WriteString(w, string(m[:]))
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var t block.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("Error: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
		}
		if !t.Validate() {
			log.Printf("Error: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		signature := utils.SignatureFromString(*t.Signature)
		bc := bcs.GetBlockchain()
		isCreated := bc.CreateTransaction(*t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, *t.Value, publicKey, signature)
		w.Header().Add("Content-Type", "application/json")
		var m []byte
		if !isCreated {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.JsonStatus("fail")
		} else {
			w.WriteHeader(http.StatusCreated)
			m = utils.JsonStatus("success")
		}
		io.WriteString(w, string(m))
	default:
		log.Println("Error")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Mine(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		isMined := bc.Mining()

		var m []byte
		if !isMined {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.JsonStatus("fail")
		} else {
			m = utils.JsonStatus("success")
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m))
	default:
		log.Println("ERROR a")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) StartMine(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		bc.StartMining()

		m := utils.JsonStatus("success")
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m))
	default:
		log.Println("ERROR a")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Amount(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		blockchainAddress := r.URL.Query().Get("blockchain_address")
		amount := bcs.GetBlockchain().CalculateTotalAmount(blockchainAddress)
		ar := &block.AmountResponse{Amount: amount}
		m, _ := ar.MarshalJSON()
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m[:]))
	default:
		log.Println("ERROR HTTP Invalid")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Run() {
	bcs.GetBlockchain().Run()
	
	http.HandleFunc("/", bcs.GetChain)
	http.HandleFunc("/transactions", bcs.Transaction)
	http.HandleFunc("/mine", bcs.Mine)
	http.HandleFunc("/mine/start", bcs.StartMine)
	http.HandleFunc("/amount", bcs.Amount)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(bcs.Port())), nil))
}
