package main

import (
	"fmt"
	"net/http"
	"text/template"
	"time"

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

	go node1.StartServer()
	go node2.StartServer()
	go node3.StartServer()

	// Keep the main goroutine alive

	// Modify the AdminProcessCertificateHandler to use the P2P network
	handler.SetNode(node1) // Use node1 for demonstration

	http.HandleFunc("/", handler.HomeHandler)
	http.HandleFunc("/register", handler.RegisterStudentHandler)
	http.HandleFunc("/login", handler.LoginStudent)
	http.HandleFunc("/student-dashboard", handler.StudentDashboardHandler)
	http.HandleFunc("/request-certificate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmpl, _ := template.ParseFiles("templates/student_request.html")
			tmpl.Execute(w, nil)
		} else if r.Method == http.MethodPost {
			name := r.FormValue("name")
			regNo := r.FormValue("regno")
			course := r.FormValue("course")

			request := p2p.CertificateRequest{
				ID:        registration.GenerateUniqueID(), // Implement this function to generate a unique integer ID
				Name:      name,
				RegNo:     regNo,
				Course:    course,
				CreatedAt: time.Now().Format(time.RFC3339),
				Status:    "Pending",
			}

			// Add request to all nodes
			node1.AddRequest(request)
			node2.AddRequest(request)
			node3.AddRequest(request)

			// Broadcast request to all nodes
			node1.BroadcastRequest(request)

			http.Redirect(w, r, "/request-success", http.StatusSeeOther)
		}
	})
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
	select {}
}
