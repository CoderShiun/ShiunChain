package blockchain

import (
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
)

const (
	dbPath = "/home/shiun/tmp/blocks"
	// to verify if our blockchain database is exists
	dbFile = "/home/shiun/tmp/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	//Blocks []*Block
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database *badger.DB
}

func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func ContiuneBlockChain(address string) *BlockChain {
	if DBexists() == false {
		fmt.Println("No exists blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	//for store the key and matadata
	opts := badger.DefaultOptions(dbPath)
	//for store the value
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}

		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return err
		})
		if err != nil {
			log.Panic(err)
		}

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	chain := BlockChain{lastHash, db}
	return &chain
}

func InitBlockChain(address string) *BlockChain {
	//return &BlockChain{[]*Block{Genesis()}}

	var lastHash []byte

	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	//for store the key and matadata
	opts := badger.DefaultOptions(dbPath)
	//for store the value
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		/*if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.WithField("Cannot save the data", err)
			}

			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash

			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			if err != nil {
				log.WithError(err)
			}

			err = item.Value(func(val []byte) error {
				lastHash = append([]byte{}, val...)
				return nil
			})
			if err != nil {
				log.WithError(err)
			}

			return err
		}*/

		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = txn.Set([]byte("lh"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		lastHash = genesis.Hash

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.WithField("Cannot get the last hash from DB", err)
		}

		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil{
			log.Panic(err)
		}

		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	if err != nil {
		log.WithField("Cannot update DB", err)
	}

	/*
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	//prevBlock.Hash = prevHash
	new := CreateBlock(data, prevBlock.Hash)
	chain.Blocks = append(chain.Blocks, new)
	*/
}

func (chain * BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	var encodedBlock []byte

	err := 	iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		if err != nil {
			log.WithField("Cannot get the current block", err)
		}

		err = item.Value(func(val []byte) error {
			encodedBlock = append([]byte{}, val...)
			block = DeSerialize(encodedBlock)
			return err
		})
		if err != nil {
			log.WithField("Cannot deserialize block", err)
		}

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	iter.CurrentHash = block.PrevHash

	return block
}

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	// key is string and value is integer
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			//make a label, so it can break from the label instead of whole loop
			Outputs:
			for outIdx, out := range tx.Outputes {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				//iter though all the tx inputs and find the outputs that is referenced by input
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTxs
}

//for finding the unspend tx output
func (chain * BlockChain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputes {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}