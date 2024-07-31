package registration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"student-certificate-validation/blockchain"
)

type Student struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	RegNo    string `json:"regno"`
	Password string `json:"password"`
}

type Admin struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type CertificateRequest struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	RegNo  string `json:"regno"`
	Course string `json:"course"`
	Status string `json:"status"`
}

// var students []Student
// var admins []Admin
// var requests []CertificateRequest
// var certificates *blockchain.Certificate

func LoadStudents() ([]Student, error) {
	file, err := os.ReadFile("students.json")
	if err != nil {
		return nil, err
	}
	var data []Student
	err = json.Unmarshal(file, &data)
	return data, err
}

func LoadRequests() ([]CertificateRequest, error) {
	file, err := ioutil.ReadFile("requests.json")
	if err != nil {
		return nil, err
	}
	var data []CertificateRequest
	err = json.Unmarshal(file, &data)
	return data, err
}

func LoadAdmins() ([]Admin, error) {
	file, err := ioutil.ReadFile("admins.json")
	if err != nil {
		return nil, err
	}
	var data []Admin
	err = json.Unmarshal(file, &data)
	return data, err
}

func SaveRequests(requests []CertificateRequest) error {
	file, err := json.MarshalIndent(requests, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("requests.json", file, 0644)
}

func SaveCertificates(certificates *blockchain.Certificate) error {
	file, err := json.MarshalIndent(certificates, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("certificates.json", file, 0644)
}
