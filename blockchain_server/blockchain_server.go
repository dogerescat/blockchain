package main

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/oniwa-shuto/blockchain/block"
	"github.com/oniwa-shuto/blockchain/wallet"
)

var chash map[string] *block.BlockChain = make(map[string]*block.BlockChain)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}

func(bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}

func(bcs *BlockchainServer) GetBlockchain() *block.BlockChain {
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
		log.Printf("ERROR: Invalid HTTP Method")
	}
}
func (bcs *BlockchainServer) Run() {
	http.HandleFunc("/", bcs.GetChain)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(bcs.Port())), nil))
}