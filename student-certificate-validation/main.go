package main

import (
	"fmt"
	"net/http"

	"student-certificate-validation/blockchain"
	"student-certificate-validation/handler"
	"student-certificate-validation/p2p"
	"student-certificate-validation/registration"
)

func main() {
	fs := http.FileServer(http.Dir("templates"))
	http.Handle("/templates/", http.StripPrefix("/templates/", fs))

	// Load students, requests, and admins
	students, err := registration.LoadStudents()
	if err != nil {
		fmt.Println("ERROR LOADING STUDENTS:", err)
		return
	}

	requests, err := registration.LoadRequests()
	if err != nil {
		fmt.Println("ERROR LOADING REQUESTS:", err)
		return
	}

	admins, err := registration.LoadAdmins()
	if err != nil {
		fmt.Println("ERROR LOADING ADMINS:", err)
		return
	}

	// Initialize Blockchain
	bc := blockchain.Blockchain{}
	if err := bc.LoadBlockchain(); err != nil {
		fmt.Println("ERROR LOADING BLOCKCHAIN:", err)
		return
	}

	// Set up HTTP handlers
	http.HandleFunc("/", handler.HomeHandler)
	http.HandleFunc("/register", handler.RegisterStudentHandler)
	http.HandleFunc("/login", handler.LoginStudent)
	http.HandleFunc("/student-dashboard", handler.StudentDashboardHandler)
	http.HandleFunc("/request-certificate", handler.StudentCertificateRequestHandler)
	http.HandleFunc("/view-download", handler.CertificateHandler)
	http.HandleFunc("/view-request", handler.ViewRequestHandler)
	http.HandleFunc("/admin-dashboard", handler.AdminDashboardHandler)
	http.HandleFunc("/admin-login", handler.AdminLoginHandler)
	http.HandleFunc("/admin-registration", handler.AdminRegistrationHandler)
	http.HandleFunc("/process-certificate", handler.AdminProcessCertificateHandler)
	http.HandleFunc("/download-certificate", handler.DownloadCertificateHandler)

	// Start P2P server
	go p2p.StartServer(":8080", &bc)

	// Start Periodic Sync
	go p2p.PeriodicSync(&bc, []p2p.Node{
		{URL: "http://localhost:8080"},
	})

	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
