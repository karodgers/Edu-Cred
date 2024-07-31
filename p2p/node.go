package p2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"student-certificate-validation/blockchain"
	"sync"
	"time"
)

type CertificateRequest struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	RegNo     string `json:"regno"`
	Course    string `json:"course"`
	CreatedAt string `json:"created_at"`
	Status    string `json:"status"`
}

type Node struct {
	ID            string
	Address       string
	Peers         map[string]string
	Blockchain    *blockchain.Blockchain
	mu            sync.Mutex
	server        *http.Server
	DataDir       string
	Requests      []CertificateRequest
	Templates     *template.Template
	PendingBlocks map[string]*blockchain.Certificate
	Approvals     map[string]map[string]bool
}

func (n *Node) AddRequest(request CertificateRequest) {
	n.Requests = append(n.Requests, request)
}

func (n *Node) GetDataFilePath(filename string) string {
	return filepath.Join(n.DataDir, filename)
}

func (n *Node) rootHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		NodeID   string
		Requests []CertificateRequest
	}{
		NodeID:   n.ID,
		Requests: n.GetPendingRequests(),
	}
	err := n.Templates.ExecuteTemplate(w, "node_dashboard.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (n *Node) GetPendingRequests() []CertificateRequest {
	var pendingRequests []CertificateRequest
	for _, req := range n.Requests {
		if req.Status == "Pending" {
			pendingRequests = append(pendingRequests, req)
		}
	}
	return pendingRequests
}

func (n *Node) BroadcastRequest(request CertificateRequest) {
	for _, peerAddr := range n.Peers {
		go func(addr string) {
			url := fmt.Sprintf("http://%s/receive-request", addr)
			jsonData, _ := json.Marshal(request)
			_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("Error broadcasting request to %s: %v\n", addr, err)
			}
		}(peerAddr)
	}
}

func NewNode(id string, address string, dataDir string) *Node {
	node := &Node{
		ID:         id,
		Address:    address,
		Peers:      make(map[string]string),
		Blockchain: blockchain.NewBlockchain(),
		DataDir:    dataDir,
		Requests:   []CertificateRequest{},
	}

	// Parse templates
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	node.Templates = templates
	node.PendingBlocks = make(map[string]*blockchain.Certificate)
	node.Approvals = make(map[string]map[string]bool)
	return node
}
func (n *Node) ProposeBlock(cert string) {
	prevCert := n.Blockchain.Certificates[len(n.Blockchain.Certificates)-1]
	newBlock := &blockchain.Certificate{
		ID:        prevCert.ID + 1,
		Name:      cert,
		PrevHash:  prevCert.Hash,
		TimeStamp: time.Now().String(),
	}
	newBlock.Hash = blockchain.GenerateHash(newBlock)

	n.PendingBlocks[newBlock.Hash] = newBlock
	n.Approvals[newBlock.Hash] = make(map[string]bool)
	n.Approvals[newBlock.Hash][n.ID] = true // Self-approve

	n.BroadcastBlockProposal(newBlock)
}
func (n *Node) ApproveBlock(blockHash string) {
	n.Approvals[blockHash][n.ID] = true

	// Check if we have majority approval
	if len(n.Approvals[blockHash]) > len(n.Peers)/2 {
		n.Blockchain.AddBlock(n.PendingBlocks[blockHash].Name)
		delete(n.PendingBlocks, blockHash)
		delete(n.Approvals, blockHash)
	}
}

func (n *Node) BroadcastBlockProposal(block *blockchain.Certificate) {
	for _, peerAddr := range n.Peers {
		go func(addr string) {
			url := fmt.Sprintf("http://%s/propose-block", addr)
			jsonData, _ := json.Marshal(block)
			_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("Error broadcasting block proposal to %s: %v\n", addr, err)
			}
		}(peerAddr)
	}
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", n.rootHandler)
	mux.HandleFunc("/request-certificate", n.requestCertificateHandler)
	mux.HandleFunc("/propose-block", n.proposeBlockHandler)
	mux.HandleFunc("/approve-block", n.approveBlockHandler)
	n.server = &http.Server{
		Addr:    n.Address,
		Handler: mux,
	}

	log.Printf("Node %s listening on %s\n", n.ID, n.Address)
	if err := n.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", n.Address, err)
	}
}
func (n *Node) proposeBlockHandler(w http.ResponseWriter, r *http.Request) {
	var block blockchain.Certificate
	err := json.NewDecoder(r.Body).Decode(&block)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n.PendingBlocks[block.Hash] = &block
	n.Approvals[block.Hash] = make(map[string]bool)
	n.ApproveBlock(block.Hash)
}

func (n *Node) approveBlockHandler(w http.ResponseWriter, r *http.Request) {
	blockHash := r.URL.Query().Get("hash")
	if blockHash == "" {
		http.Error(w, "Missing block hash", http.StatusBadRequest)
		return
	}

	n.ApproveBlock(blockHash)
}
func (n *Node) requestCertificateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := n.Templates.ExecuteTemplate(w, "student_request.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		regNo := r.FormValue("regno")
		course := r.FormValue("course")

		request := CertificateRequest{
			ID:        len(n.Requests) + 1,
			Name:      name,
			RegNo:     regNo,
			Course:    course,
			CreatedAt: time.Now().Format(time.RFC3339),
			Status:    "Pending",
		}

		n.AddRequest(request)
		n.BroadcastRequest(request)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

