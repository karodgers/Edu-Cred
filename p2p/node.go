package p2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"student-certificate-validation/blockchain"
	"sync"
	"time"
)

type Node struct {
	ID         string
	Address    string
	Peers      map[string]string
	Blockchain *blockchain.Blockchain
	mu         sync.Mutex
	server     *http.Server
	DataDir    string
}

func (n *Node) GetDataFilePath(filename string) string {
	return filepath.Join(n.DataDir, filename)
}

func (n *Node) rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "P2P Node %s is running", n.ID)
}

func NewNode(id string, address string, dataDir string) *Node {
	node := &Node{
		ID:         id,
		Address:    address,
		Peers:      make(map[string]string),
		Blockchain: blockchain.NewBlockchain(),
		DataDir:    dataDir,
	}

	// Set the node in the blockchain
	node.Blockchain.SetNode(node)

	mux := http.NewServeMux()
	mux.HandleFunc("/receive-block", node.receiveBlockHandler)

	// Add the root handler
	mux.HandleFunc("/", node.rootHandler)

	node.server = &http.Server{
		Addr:    address,
		Handler: mux,
	}

	return node
}

func (n *Node) AddPeer(id, address string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Peers[id] = address
}

func (n *Node) BroadcastBlock(cert blockchain.Certificate) {
	for _, peerAddr := range n.Peers {
		go func(addr string) {
			url := fmt.Sprintf("http://%s/receive-block", addr)
			jsonData, _ := json.Marshal(cert)
			_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("Error broadcasting to %s: %v\n", addr, err)
			}
		}(peerAddr)
	}
}

func (n *Node) StartServer() {
	go func() {
		if err := n.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Node %s listen error: %s\n", n.ID, err)
		}
	}()
	fmt.Printf("Node %s started on %s\n", n.ID, n.Address)
}

func (n *Node) receiveBlockHandler(w http.ResponseWriter, r *http.Request) {
	var cert blockchain.Certificate
	err := json.NewDecoder(r.Body).Decode(&cert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n.Blockchain.AddBlock(cert.Name)
	fmt.Printf("Node %s received a new block: %v\n", n.ID, cert)
}

func (n *Node) MineBlock(cert string) {
	// Simple Proof of Work
	for {
		n.Blockchain.AddBlock(cert)
		lastBlock := n.Blockchain.Certificates[len(n.Blockchain.Certificates)-1]
		if lastBlock.Hash[:4] == "0000" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	n.BroadcastBlock(n.Blockchain.Certificates[len(n.Blockchain.Certificates)-1])
}
