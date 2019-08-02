package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	log "github.com/sirupsen/logrus"
)

type Block struct {
	Hash []byte
	//Data     []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

/*
//create hash based on data and prevHash
func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}*/

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

//create one block
func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{[]byte{}, txs, prevHash, 0}

	/*
	//insert Block.Hash
	block.DeriveHash()
	*/

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return res.Bytes()
}

func DeSerialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
