package main

import (
	"fmt"
	"net/http"
	"student-certificate-validation/handler"
	"student-certificate-validation/p2p"
	"student-certificate-validation/registration"
)

func main() {
	fs := http.FileServer(http.Dir("templates"))
	http.Handle("/templates/", http.StripPrefix("/templates/", fs))

	// Load students from the file
	if err := registration.LoadStudents(); err != nil {
		fmt.Println("ERROR LOADING STUDENTS:", err)
		return
	}

	// Load certificate requests from the file
	if err := registration.LoadRequests(); err != nil {
		fmt.Println("ERROR LOADING REQUESTS:", err)
		return
	}

	// Load admins from the file
	if err := registration.LoadAdmins(); err != nil {
		fmt.Println("ERROR LOADING ADMINS:", err)
		return
	}

	// Create and start P2P nodes
	node1 := p2p.NewNode("node1", "localhost:8001", "./data/node1")
	node2 := p2p.NewNode("node2", "localhost:8002", "./data/node2")
	node3 := p2p.NewNode("node3", "localhost:8003", "./data/node3")

	node1.AddPeer("node2", "localhost:8002")
	node1.AddPeer("node3", "localhost:8003")
	node2.AddPeer("node1", "localhost:8001")
	node2.AddPeer("node3", "localhost:8003")
	node3.AddPeer("node1", "localhost:8001")
	node3.AddPeer("node2", "localhost:8002")

	node1.StartServer()
	node2.StartServer()
	node3.StartServer()

	// Modify the AdminProcessCertificateHandler to use the P2P network
	handler.SetNode(node1) // Use node1 for demonstration

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

	fmt.Println("Main server running at http://localhost:1234")
	fmt.Println("P2P nodes running at:")
	fmt.Println("- http://localhost:8001")
	fmt.Println("- http://localhost:8002")
	fmt.Println("- http://localhost:8003")

	http.ListenAndServe(":1234", nil)
}
