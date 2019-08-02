package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	log "github.com/sirupsen/logrus"

)

/*Steps:
1. Take the data from the block
2. Create a counter (nonce) which starts at 0
3. Create a hash of the data plus the counter
4. Check the hash to see if it meets a set of requirements
Requirements: The first few bytes must contain 0s*/

const Difficulty = 12

type ProofOfWork struct {
	Block *Block
	Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	//难度：target=10的18次方（即要求计算出的hash值小于这个target）
	target.Lsh(target, uint(256-Difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		}, []byte{},)

	return data
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Error(err)
	}

	return buff.Bytes()
}

func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)

		intHash.SetBytes(hash[:])
		//ToDo: check what is intHash

		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce ++
		}
	}
	fmt.Println()

	return nonce, hash[:]
}