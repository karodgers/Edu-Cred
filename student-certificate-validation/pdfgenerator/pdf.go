package pdfgenerator

import (
	"fmt"
	"student-certificate-validation/blockchain"

	"github.com/signintech/gopdf"
)

func GeneratePDF(certificate blockchain.Certificate, hash string) (string, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	pdf.AddPage()

	pdf.AddTTFFont("arial", "arial.ttf")
	pdf.SetFont("arial", "", 14)

	pdf.Cell(nil, fmt.Sprintf("Certificate ID: %d", certificate.ID))
	pdf.Cell(nil, fmt.Sprintf("Name: %s", certificate.Name))
	pdf.Cell(nil, fmt.Sprintf("Registration No: %s", certificate.RegNo))
	pdf.Cell(nil, fmt.Sprintf("Course: %s", certificate.Course))
	pdf.Cell(nil, fmt.Sprintf("Date: %s", certificate.CreatedAt))
	pdf.Cell(nil, fmt.Sprintf("Hash: %s", hash))

	filePath := fmt.Sprintf("certificates/certificate_%d.pdf", certificate.ID)
	err := pdf.WritePdf(filePath)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
