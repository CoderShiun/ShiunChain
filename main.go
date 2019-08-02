package main

import (
	"ShiunChain/blockchain"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	log "github.com/sirupsen/logrus"
	"time"
)

type CommandLine struct {
	blockchain *blockchain.BlockChain
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" add -block BLOCK_DATA - add a block to the chain")
	fmt.Println(" print - Prints the block in the chain")
}

func (cli * CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli * CommandLine) addBlock(data string){
	cli.blockchain.AddBlock(data)
	fmt.Println("Added Block!")
}

func (cli * CommandLine) printChain() {
	iter := cli.blockchain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("POW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli * CommandLine) run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.WithField("Cannot add block", err)
		}

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.WithField("There is no block to print", err)
		}

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		} else {
			cli.addBlock(*addBlockData)
		}
	}

	if printChainCmd.Parsed() {
		cli.printUsage()
	}
}

func main(){
	defer os.Exit(0)
	chain := blockchain.InitBlockChain()
	defer chain.Database.Close()

	cli := CommandLine{chain}
	cli.run()

	time.Sleep(15 * time.Second)
}