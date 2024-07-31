package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"student-certificate-validation/blockchain"
	"student-certificate-validation/p2p"
	"student-certificate-validation/pdfgenerator"
	"student-certificate-validation/registration"
	"time"
)

func AdminProcessCertificateHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	request.Mutex.Lock()
	defer requestMutex.Unlock()

	var request *registration.CertificateRequest
	for i := range requests {
		if requests[i].ID == id {
			request = &requests[i]
			break
		}
	}
	if request == nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	if request.Status != "Pending" {
		http.Error(w, "Request already processed", http.StatusBadRequest)
		return
	}

	// Update status and create a new certificate
	request.Status = "Completed"
	certificate := registration.Certificate{
		ID:        len(certificates) + 1,
		Name:      request.Name,
		RegNo:     request.RegNo,
		Course:    request.Course,
		CreatedAt: time.Now().String(),
	}

	// Add the certificate to the blockchain
	bc := blockchain.Blockchain{}
	if err := bc.LoadBlockchain(); err != nil {
		http.Error(w, fmt.Sprintf("Error loading blockchain: %v", err), http.StatusInternalServerError)
		return
	}

	err = bc.AddBlock(certificate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error adding block to blockchain: %v", err), http.StatusInternalServerError)
		return
	}

	certificate.Hash = bc.Blocks[len(bc.Blocks)-1].Hash

	// Generate the PDF
	filePath, err := pdfgenerator.GeneratePDF(certificate, certificate.Hash)
	if err != nil {
		http.Error(w, "Error generating PDF", http.StatusInternalServerError)
		return
	}

	// Save the updated requests and certificates
	if err := registration.SaveRequests(requests); err != nil {
		http.Error(w, "Error saving requests", http.StatusInternalServerError)
		return
	}

	if err := registration.SaveCertificates(certificates); err != nil {
		http.Error(w, "Error saving certificates", http.StatusInternalServerError)
		return
	}

	if err := bc.saveBlockchain(); err != nil {
		http.Error(w, fmt.Sprintf("Error saving blockchain: %v", err), http.StatusInternalServerError)
		return
	}

	// Notify peers about the new certificate
	peers := []p2p.Node{
		{URL: "http://localhost:5001"},
	}

	for _, peer := range peers {
		if err := peer.SyncBlockchain(&bc); err != nil {
			fmt.Printf("Failed to sync with peer %s: %v\n", peer.URL, err)
		}
	}

	// Redirect to download the generated PDF
	http.Redirect(w, r, "/download-certificate?file="+filePath, http.StatusSeeOther)
}
