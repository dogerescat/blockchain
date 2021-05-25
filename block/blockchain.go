package block

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/oniwa-shuto/blockchain/utils"
)

const (
	MAINING_DIFFICULTY = 3
	MAINING_SENDER     = "THE BLOCKCHAIN"
	MAINING_REWORD     = 1.0
)

type Block struct {
	timestamp    int64
	nonce        int
	previousHash [32]byte
	transactions []*Transaction
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*Transaction) *Block {
	return &Block{
		nonce:        nonce,
		previousHash: previousHash,
		timestamp:    time.Now().UnixNano(),
		transactions: transactions,
	}
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

//methodを上書きMarshalJSON 大文字じゃないと上書きできない
func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"noce"`
		PreviousHash string       `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
	})
}

// func Print(bc *BlockChain) {
// 	for i, block := range bc.chain {
// 		fmt.Printf("Chain %d \n", i)
// 		block.Print()
// 	}
// }

func (b *Block) Print() {
	fmt.Printf("nonce: %d \n", b.nonce)
	fmt.Printf("timestamp: %d \n", b.timestamp)
	fmt.Printf("previous_hash: %x \n", b.previousHash)
	for _, t := range b.transactions {
		t.Print()
	}
}

type BlockChain struct {
	transactionPool   []*Transaction
	chain             []*Block
	blockchainAddress string
	port              uint16
}

func NewBlockChain(blockchainAddress string, port uint16) *BlockChain {
	b := &Block{}

	bc := new(BlockChain)
	bc.blockchainAddress = blockchainAddress
	bc.CreateBlock(0, b.Hash())
	bc.port = port
	return bc
}

func (bc *BlockChain) MarshalJSON() ([]byte, error){
	return json.Marshal(struct {
		Blocks []*Block `json:"chains"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *BlockChain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}
	return b
}

func (bc *BlockChain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *BlockChain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	fmt.Printf("%s\n", strings.Repeat("*", 25))
}

func (bc *BlockChain) AddTransaction(sender string, recipient string, value float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)
	if sender == MAINING_SENDER {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}
	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		/*
			if bc.CalculateTotalAmount(sender) < value {
				log.Println("ERROR: Not enogh balance in wallet")
				return false
			}
		*/
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("Error: Verify Transaction")
	}
	return false
}

func (bc *BlockChain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)

}

func (bc *BlockChain) CopyTransaction() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions,
			NewTransaction(t.senderBlockchainAddress, t.recipientBlockAddress, t.value))
	}
	return transactions
}

func (bc *BlockChain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{0, nonce, previousHash, transactions}
	gussHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	return gussHashStr[:difficulty] == zeros
}

func (bc *BlockChain) ProofOfWork() int {
	transactions := bc.CopyTransaction()
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MAINING_DIFFICULTY) {
		nonce += 1
	}
	return nonce
}

func (bc *BlockChain) Maining() bool {
	bc.AddTransaction(MAINING_SENDER, bc.blockchainAddress, MAINING_REWORD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	log.Println("action=maining, status=success")
	return true
}

func (bc *BlockChain) CalculateTotalAmount(blockchainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, b := range bc.chain {
		for _, t := range b.transactions {
			value := t.value
			if blockchainAddress == t.recipientBlockAddress {
				totalAmount += value
			}
			if blockchainAddress == t.senderBlockchainAddress {
				totalAmount -= value
			}
		}
	}
	return totalAmount
}

type Transaction struct {
	senderBlockchainAddress string
	recipientBlockAddress   string
	value                   float32
}

func NewTransaction(sender string, recipient string, value float32) *Transaction {
	return &Transaction{sender, recipient, value}
}

func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf("sender_blockchain_address      %s\n", t.senderBlockchainAddress)
	fmt.Printf("recipient_blockchain_address   %s\n", t.recipientBlockAddress)
	fmt.Printf("value                          %.1f\n", t.value)

}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockAddress,
		Value:     t.value,
	})

}
