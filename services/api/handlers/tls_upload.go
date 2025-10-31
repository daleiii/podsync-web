package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	maxUploadSize = 10 * 1024 * 1024 // 10 MB
	tlsCertsDir   = "./certs"        // Directory to store uploaded certificates
)

type TLSUploadResponse struct {
	CertificatePath string `json:"certificate_path"`
	KeyFilePath     string `json:"key_file_path"`
	Message         string `json:"message"`
}

// HandleTLSUpload handles uploading TLS certificate and key files
func HandleTLSUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.WithError(err).Error("failed to parse multipart form")
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Create certs directory if it doesn't exist
	if err := os.MkdirAll(tlsCertsDir, 0755); err != nil {
		log.WithError(err).Error("failed to create certs directory")
		http.Error(w, "Failed to create certificates directory", http.StatusInternalServerError)
		return
	}

	response := TLSUploadResponse{}

	// Handle certificate file upload
	certFile, certHeader, err := r.FormFile("certificate")
	if err == nil {
		defer certFile.Close()

		// Validate file extension
		ext := filepath.Ext(certHeader.Filename)
		if ext != ".pem" && ext != ".crt" && ext != ".cer" {
			http.Error(w, "Certificate file must be .pem, .crt, or .cer", http.StatusBadRequest)
			return
		}

		// Save certificate file
		certPath := filepath.Join(tlsCertsDir, "server.crt")
		if err := saveUploadedFile(certFile, certPath); err != nil {
			log.WithError(err).Error("failed to save certificate file")
			http.Error(w, "Failed to save certificate file", http.StatusInternalServerError)
			return
		}

		// Get absolute path
		absPath, err := filepath.Abs(certPath)
		if err != nil {
			log.WithError(err).Error("failed to get absolute path for certificate")
			http.Error(w, "Failed to process certificate path", http.StatusInternalServerError)
			return
		}

		response.CertificatePath = absPath
		log.Infof("Certificate uploaded to %s", absPath)
	} else if err != http.ErrMissingFile {
		log.WithError(err).Error("failed to read certificate file")
		http.Error(w, "Failed to read certificate file", http.StatusBadRequest)
		return
	}

	// Handle key file upload
	keyFile, keyHeader, err := r.FormFile("key")
	if err == nil {
		defer keyFile.Close()

		// Validate file extension
		ext := filepath.Ext(keyHeader.Filename)
		if ext != ".pem" && ext != ".key" {
			http.Error(w, "Key file must be .pem or .key", http.StatusBadRequest)
			return
		}

		// Save key file with restricted permissions
		keyPath := filepath.Join(tlsCertsDir, "server.key")
		if err := saveUploadedFile(keyFile, keyPath); err != nil {
			log.WithError(err).Error("failed to save key file")
			http.Error(w, "Failed to save key file", http.StatusInternalServerError)
			return
		}

		// Set restrictive permissions on key file
		if err := os.Chmod(keyPath, 0600); err != nil {
			log.WithError(err).Error("failed to set permissions on key file")
			http.Error(w, "Failed to secure key file", http.StatusInternalServerError)
			return
		}

		// Get absolute path
		absPath, err := filepath.Abs(keyPath)
		if err != nil {
			log.WithError(err).Error("failed to get absolute path for key")
			http.Error(w, "Failed to process key path", http.StatusInternalServerError)
			return
		}

		response.KeyFilePath = absPath
		log.Infof("Key file uploaded to %s", absPath)
	} else if err != http.ErrMissingFile {
		log.WithError(err).Error("failed to read key file")
		http.Error(w, "Failed to read key file", http.StatusBadRequest)
		return
	}

	// Check if at least one file was uploaded
	if response.CertificatePath == "" && response.KeyFilePath == "" {
		http.Error(w, "No certificate or key file provided", http.StatusBadRequest)
		return
	}

	response.Message = "TLS files uploaded successfully"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func saveUploadedFile(src io.Reader, dst string) error {
	// Create destination file
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Copy file contents
	if _, err := io.Copy(out, src); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
