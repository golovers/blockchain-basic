package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

var difficulty = 4

// Block block in the blockchain
type Block struct {
	Timestamp  time.Time
	PrevHash   []byte
	Difficulty int
	Nonce      int

	Data []byte
}

// Blockchain the main chain
type Blockchain struct {
	chains []*Block
	rschan chan *Block
}

func newBlock(data []byte, prevHash []byte) *Block {
	b := &Block{
		Data:       data,
		PrevHash:   prevHash,
		Timestamp:  time.Now(),
		Difficulty: difficulty,
	}
	b.Nonce = proofOfWork(b)
	return b
}

// CalHash return hash of the block
func (block *Block) CalHash() []byte {
	h := sha256.New()
	b := bytes.Join([][]byte{toBytes(block.Timestamp), block.PrevHash, block.Data, toBytes(block.Difficulty), toBytes(block.Nonce)}, []byte{})
	h.Write(b)
	return h.Sum(nil)
}

// Print print block info to console
func (block *Block) Print() {
	fmt.Printf("Hash: %s\n", hex.EncodeToString(block.CalHash()))
	fmt.Printf("PrevHash: %s\n", hex.EncodeToString(block.PrevHash))
	fmt.Println("Nonce:", block.Nonce)
	fmt.Printf("Data: %s\n", string(block.Data))
	fmt.Println("POW:", validBlock(block))
	fmt.Println()
}

func hashit(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func genesis() *Block {
	return newBlock([]byte("I'm genesis block"), []byte{})
}

// NewBlockchain return a new blockchain with genesis block inside
func NewBlockchain() *Blockchain {
	return &Blockchain{
		chains: []*Block{genesis()},
		rschan: make(chan *Block),
	}
}

// Chains return the blockchain for iterator
func (blockchain *Blockchain) Chains() []*Block {
	return blockchain.chains
}

// AddBlock add a new block into the blockchain
func (blockchain *Blockchain) AddBlock(data []byte) {
	chains := blockchain.Chains()
	prevBlock := chains[len(chains)-1]
	b := newBlock(data, prevBlock.CalHash())
	if !validBlock(b) {
		fmt.Println("ERROR: Fake block....")
	}
	blockchain.chains = append(blockchain.chains, b)
	blockchain.rschan <- b
}

// proofOfWork find nonce so that hash of data of the block and nonce has #Difficulty leading zero
func proofOfWork(b *Block) int {
	blockData := func(nonce int) []byte {
		return bytes.Join([][]byte{
			toBytes(b.Timestamp),
			b.PrevHash,
			b.Data,
			toBytes(b.Difficulty),
			toBytes(nonce),
		}, []byte{})
	}
	// find nonce
	nonce := 0
	prefix := strings.Repeat("0", difficulty)
	for {
		hashValue := hex.EncodeToString(hashit(blockData(nonce)))
		if strings.HasPrefix(hashValue, prefix) {
			return nonce
		}
		nonce++
	}
}

// Validate validBlock the proof of work
func validBlock(block *Block) bool {
	prefix := strings.Repeat("0", block.Difficulty)
	hashValue := hex.EncodeToString(block.CalHash())
	return strings.HasPrefix(hashValue, prefix)
}

func toBytes(v interface{}) []byte {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(v)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func main() {
	fmt.Println("---------Basic blockchain----------")
	blockchain := NewBlockchain()
	blockchain.Chains()[0].Print()
	go func() {
		for {
			select {
			case b := <-blockchain.rschan:
				b.Print()
			}
		}
	}()

	i := 1
	for i <= 5 {
		blockchain.AddBlock([]byte(fmt.Sprintf("Block %v", i)))
		i++
	}
}
