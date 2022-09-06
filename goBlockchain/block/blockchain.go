package block

import (
	"crypto/ecdsa"
	"crypto/sha256" //used to implement  the SHA224 and SHA256 hash algorithms
	"encoding/json"
	"fmt"
	"goblockchain/utils"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	MINING_DIFFICULTY = 3
	MINING_SENDER = "THE BLOCKCHAIN"
	MINING_REWARD = 1.0
	MINING_TIMER_SEC = 20
)

type Block struct {
	Timestamp    int64         `json:"timestamp"` //The fields inside struct need to be in uppercase if we are using the struct to make a json
    Nonce        int            `json:"nonce"`
	PreviousHash [32]byte       `json:"previous_hash"`
	Transactions []*Transaction `json:"transactions"`
}

type Blockchain struct {
	transactionPool []*Transaction
	chain           []*Block
	blockchainAddress string
	port            uint16
    mux             sync.Mutex
}

type Transaction struct {
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}


func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block{
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}
	return b
}

func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
	bc.CreateBlock(0,b.Hash() )
	bc.port = port
	return bc
}

func(bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:chains`
	}{
		Blocks: bc.chain,
	})
}

func (b *Block) print() {
	fmt.Printf("timestamp %d\n", b.Timestamp)
	fmt.Printf("nonce %d\n", b.Nonce)
	fmt.Printf("previous hash %s\n", b.PreviousHash)
	for _,t := range b.Transactions {
		t.Print()
	}
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n",strings.Repeat("-",25), i, strings.Repeat("-",25))  //strings.Repeat() takes two argument n & i and repeats n, i number of times
		block.print()
	}
	fmt.Printf("%sEnd of chain%s\n",strings.Repeat("=",25), strings.Repeat("=",25))
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*Transaction) *Block {
	b := new(Block) //new() method takes a type as an argument and returns a pointer to new variable of the mentioned type
    b.Timestamp = time.Now().UnixNano()
	b.Nonce = nonce
	b.PreviousHash = previousHash
	b.Transactions = transactions

    return b
}

func NewTransaction(sender string, recipient string, value float32) *Transaction{
	return &Transaction{sender, recipient, value}
}

func (bc *Blockchain) CreateTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransacted := bc.AddTransaction(sender, recipient, value, senderPublicKey, s) 
	
	//TODO:FIXME:

	return isTransacted
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)

	if sender == MINING_SENDER{
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		/*
		if bc.CalculateTotalAmount(sender) < value {
			log.Println("Error: Not enough balance in the wallet")
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

func (bc *Blockchain) VerifyTransactionSignature(senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
   m, _ := json.Marshal(t)
   h := sha256.Sum256([]byte(m))
   return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S )
}

func(bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions, NewTransaction(t.senderBlockchainAddress, t.recipientBlockchainAddress, t.value))
	}
	return transactions
}

func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool{
 zeros := strings.Repeat("0", difficulty)
 guessBlock := Block{0, nonce, previousHash,transactions}
 guessBlockHashStr := fmt.Sprintf("%x",guessBlock.Hash())
 return guessBlockHashStr[:difficulty] == zeros
}

func(bc *Blockchain) ProofOfWork() int {
  transaction := bc.CopyTransactionPool()
  previousHash := bc.LastBlock().Hash()
  nonce := 0
  for !bc.ValidProof(nonce, previousHash, transaction, MINING_DIFFICULTY) { //the cycle will continue until the function returns true
     nonce += 1
  }
  return nonce
}

func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()
	defer bc.mux.Unlock()

    if len(bc.transactionPool) == 0 {
		return false
	}

	bc.AddTransaction(MINING_SENDER, bc.blockchainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce,previousHash)
	log.Println("action=mining status=success")
	return true
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second * MINING_TIMER_SEC, bc.StartMining)
}

func (bc *Blockchain) CalculateTotalAmount(blockchainAddress string) float32 {  //function will return thr total amount of token in possession of the user having the blockchain address passed in the argument.
	var totalAmount float32 = 0.0
	for _,b := range bc.chain {
		for _,t := range b.Transactions {
			value := t.value
			if blockchainAddress == t.recipientBlockchainAddress {
				totalAmount += value
			}
			if blockchainAddress == t.senderBlockchainAddress {
				totalAmount -= value
			}
		}
	
	}
	return totalAmount
}

func(t *Transaction) Print() {
	fmt.Printf("%s\n",strings.Repeat("-",30))
	fmt.Printf("sender_blockchain_address    %s\n",   t.senderBlockchainAddress)
	fmt.Printf("recipient_blockchain_address %s\n",   t.recipientBlockchainAddress)
	fmt.Printf("transaction_value            %.1f\n", t.value)
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct{
		Sender string    `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value float32      `json:"value"`
	}{
		Sender: t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value: t.value,
	})
}




func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct { //The fields inside struct need to be in uppercase if we are using the struct to make a json
		Timestamp    int64    `json:"timestamp"`
		Nonce        int      `json:"nonce"`
		PreviousHash string`json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.Timestamp,
		Nonce:        b.Nonce,
		PreviousHash: fmt.Sprintf("%x",b.PreviousHash),
		Transactions: b.Transactions,
	})
}

 

func (b *Block) Hash() [32]byte{ //Sum256() method on sha256 package returns a slice with 32 bytes
  m,_ := json.Marshal(b)
  return sha256.Sum256(m)
}

func(bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

type TransactionRequest struct {
  SenderBlockchainAddress *string `json:"sender_blockchain_address"`
  RecipientBlockchainAddress *string `json:"recipient_blockchain_address"`
  SenderPublicKey *string `json:"sender_public_key"`
  Value *float32 `json:"value"`
  Signature *string `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.Signature == nil || tr.SenderBlockchainAddress == nil || tr.RecipientBlockchainAddress == nil || tr.SenderPublicKey == nil || tr.Value == nil {
		return false
	} 
		return true
}

type AmountResponse struct {
	Amount float32 `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct{
		Amount float32 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}