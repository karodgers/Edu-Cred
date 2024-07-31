package p2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"student-certificate-validation/blockchain"
	"time"
)

type Node struct {
	URL string
}

func (n *Node) SyncBlockchain(bc *blockchain.Blockchain) error {
	data, err := json.Marshal(bc)
	if err != nil {
		return err
	}
	resp, err := http.Post(n.URL+"/sync", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func StartServer(port string, bc *blockchain.Blockchain) {
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var newBC blockchain.Blockchain
			if err := json.NewDecoder(r.Body).Decode(&newBC); err != nil {
				http.Error(w, "Invalid data", http.StatusBadRequest)
				return
			}
			*bc = newBC
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(port, nil)
}

func PeriodicSync(bc *blockchain.Blockchain, nodes []Node) {
	for {
		for _, node := range nodes {
			if err := node.SyncBlockchain(bc); err != nil {
				fmt.Printf("Failed to sync with peer %s: %v\n", node.URL, err)
			}
		}
		time.Sleep(10 * time.Minute)
	}
}
