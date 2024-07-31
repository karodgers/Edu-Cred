package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type Certificate struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	RegNo     string `json:"regno"`
	Course    string `json:"course"`
	CreatedAt string `json:"created_at"`
	Hash      string `json:"hash"`
}

type Block struct {
	Index        int         `json:"index"`
	Timestamp    string      `json:"timestamp"`
	Certificate  Certificate `json:"certificate"`
	PreviousHash string      `json:"previous_hash"`
	Hash         string      `json:"hash"`
}

type Blockchain struct {
	Blocks []Block `json:"blocks"`
}

func (bc *Blockchain) AddBlock(certificate Certificate) error {
	var previousHash string
	if len(bc.Blocks) > 0 {
		previousHash = bc.Blocks[len(bc.Blocks)-1].Hash
	}

	block := Block{
		Index:        len(bc.Blocks) + 1,
		Timestamp:    time.Now().String(),
		Certificate:  certificate,
		PreviousHash: previousHash,
		Hash:         generateHash(len(bc.Blocks)+1, time.Now().String(), certificate, previousHash),
	}

	bc.Blocks = append(bc.Blocks, block)
	return bc.saveBlockchain()
}

func (bc *Blockchain) saveBlockchain() error {
	file, err := json.MarshalIndent(bc, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("blocks.json", file, 0644)
}

func (bc *Blockchain) LoadBlockchain() error {
	file, err := ioutil.ReadFile("blocks.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &bc)
}

func generateHash(index int, timestamp string, certificate Certificate, prevHash string) string {
	record := fmt.Sprintf("%d%s%v%s", index, timestamp, certificate, prevHash)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(record)))
}
