package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	uploadDir    = "./uploads"
	maxFileSize  = 50 * 1024 * 1024 // 50MB
	allowedTypes = "image/jpeg,image/png,image/gif,application/pdf,text/plain,application/zip"
)

type FileInfo struct {
	ID          string    `json:"id"`
	OriginalName string   `json:"original_name"`
	Size        int64     `json:"size"`
	UploadTime  time.Time `json:"upload_time"`
	MimeType    string    `json:"mime_type"`
}

var fileStore = make(map[string]FileInfo)

// securityHeadersMiddleware adds security headers to all responses
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// CORS headers for API access
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// healthCheckHandler provides health check endpoint for deployment platforms
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ok", "service": "file-sharing"}`)
}

func main() {
	// Create uploads directory
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	r := mux.NewRouter()

	// Add security middleware
	r.Use(securityHeadersMiddleware)

	// Static file serving
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Web UI routes
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/upload", uploadHandler).Methods("POST")
	r.HandleFunc("/download/{id}", downloadHandler).Methods("GET")
	r.HandleFunc("/info/{id}", fileInfoHandler).Methods("GET")

	// API routes (for curl)
	r.HandleFunc("/api/upload", apiUploadHandler).Methods("POST")
	r.HandleFunc("/api/info/{id}", apiFileInfoHandler).Methods("GET")

	// Health check endpoint for deployment
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Open your browser at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File Sharing Service</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .upload-area {
            border: 2px dashed #ddd;
            border-radius: 10px;
            padding: 40px;
            text-align: center;
            margin: 20px 0;
            transition: border-color 0.3s;
        }
        .upload-area:hover {
            border-color: #007bff;
        }
        .upload-area.dragover {
            border-color: #007bff;
            background-color: #f8f9fa;
        }
        input[type="file"] {
            margin: 20px 0;
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
            transition: background-color 0.3s;
        }
        button:hover {
            background-color: #0056b3;
        }
        .result {
            margin-top: 20px;
            padding: 15px;
            border-radius: 5px;
            display: none;
        }
        .success {
            background-color: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
        }
        .error {
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
        }
        .file-info {
            background-color: #e9ecef;
            padding: 15px;
            border-radius: 5px;
            margin-top: 15px;
        }
        .copy-btn {
            background-color: #28a745;
            font-size: 12px;
            padding: 5px 10px;
            margin-left: 10px;
        }
        .copy-btn:hover {
            background-color: #218838;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📁 File Sharing Service</h1>
        
        <div class="upload-area" id="uploadArea">
            <p>📤 Drag files here or click to select</p>
            <input type="file" id="fileInput" multiple>
        </div>
        
        <div style="text-align: center;">
            <button onclick="uploadFiles()">Upload</button>
        </div>
        
        <div id="result" class="result"></div>
    </div>

    <script>
        const uploadArea = document.getElementById('uploadArea');
        const fileInput = document.getElementById('fileInput');
        const result = document.getElementById('result');

        // 드래그 앤 드롭 이벤트
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('dragover');
        });

        uploadArea.addEventListener('dragleave', () => {
            uploadArea.classList.remove('dragover');
        });

        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
            fileInput.files = e.dataTransfer.files;
        });

        uploadArea.addEventListener('click', () => {
            fileInput.click();
        });

        async function uploadFiles() {
            const files = fileInput.files;
            if (files.length === 0) {
                showResult('Please select files.', 'error');
                return;
            }

            const formData = new FormData();
            for (let file of files) {
                formData.append('files', file);
            }

            try {
                const response = await fetch('/upload', {
                    method: 'POST',
                    body: formData
                });

                const data = await response.json();
                
                if (response.ok) {
                    let html = '<div class="file-info">';
                    data.files.forEach(file => {
                        const shareUrl = window.location.origin + '/download/' + file.id;
                        html += '<p><strong>Filename:</strong> ' + file.original_name + '</p>';
                        html += '<p><strong>Size:</strong> ' + formatFileSize(file.size) + '</p>';
                        html += '<p><strong>Share Link:</strong> ';
                        html += '<input type="text" value="' + shareUrl + '" readonly style="width: 300px; padding: 5px;">';
                        html += '<button class="copy-btn" onclick="copyToClipboard(\'' + shareUrl + '\')">Copy</button>';
                        html += '</p><hr>';
                    });
                    html += '</div>';
                    showResult(html, 'success');
                } else {
                    showResult('Upload failed: ' + data.error, 'error');
                }
            } catch (error) {
                showResult('Error during upload: ' + error.message, 'error');
            }
        }

        function showResult(message, type) {
            result.innerHTML = message;
            result.className = 'result ' + type;
            result.style.display = 'block';
        }

        function formatFileSize(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function copyToClipboard(text) {
            navigator.clipboard.writeText(text).then(() => {
                alert('Link copied to clipboard!');
            });
        }
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 파일 크기 제한
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		http.Error(w, `{"error": "File size too large (max 50MB)"}`, http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, `{"error": "No files to upload"}`, http.StatusBadRequest)
		return
	}

	var uploadedFiles []FileInfo

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, `{"error": "Cannot open file"}`, http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// 파일 검증
		if err := validateFile(fileHeader); err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		// 고유 ID 생성
		fileID := generateFileID(fileHeader.Filename)
		
		// 파일 저장
		filePath := filepath.Join(uploadDir, fileID)
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, `{"error": "Failed to save file"}`, http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, `{"error": "Failed to save file"}`, http.StatusInternalServerError)
			return
		}

		// 파일 정보 저장
		fileInfo := FileInfo{
			ID:           fileID,
			OriginalName: fileHeader.Filename,
			Size:         fileHeader.Size,
			UploadTime:   time.Now(),
			MimeType:     fileHeader.Header.Get("Content-Type"),
		}
		fileStore[fileID] = fileInfo
		uploadedFiles = append(uploadedFiles, fileInfo)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Upload successful", "files": %s}`, toJSON(uploadedFiles))
}

func apiUploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 파일 크기 제한
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		http.Error(w, `{"error": "File size too large (max 50MB)"}`, http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, `{"error": "No files to upload"}`, http.StatusBadRequest)
		return
	}

	var uploadedFiles []FileInfo

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, `{"error": "Cannot open file"}`, http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// 파일 검증
		if err := validateFile(fileHeader); err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		// 고유 ID 생성
		fileID := generateFileID(fileHeader.Filename)
		
		// 파일 저장
		filePath := filepath.Join(uploadDir, fileID)
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, `{"error": "Failed to save file"}`, http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, `{"error": "Failed to save file"}`, http.StatusInternalServerError)
			return
		}

		// 파일 정보 저장
		fileInfo := FileInfo{
			ID:           fileID,
			OriginalName: fileHeader.Filename,
			Size:         fileHeader.Size,
			UploadTime:   time.Now(),
			MimeType:     fileHeader.Header.Get("Content-Type"),
		}
		fileStore[fileID] = fileInfo
		uploadedFiles = append(uploadedFiles, fileInfo)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Upload successful", "files": %s}`, toJSON(uploadedFiles))
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	// Validate file ID format (must be hex string)
	if len(fileID) != 32 || !isHexString(fileID) {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	fileInfo, exists := fileStore[fileID]
	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Ensure the file path is within the upload directory (prevent path traversal)
	filePath := filepath.Join(uploadDir, fileID)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "Invalid file path", http.StatusInternalServerError)
		return
	}
	
	absUploadDir, _ := filepath.Abs(uploadDir)
	if !strings.HasPrefix(absPath, absUploadDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File does not exist", http.StatusNotFound)
		return
	}

	// Set secure headers
	w.Header().Set("Content-Disposition", "attachment; filename=\""+sanitizeFilename(fileInfo.OriginalName)+"\"")
	w.Header().Set("Content-Type", fileInfo.MimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, filePath)
}

func fileInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	// Validate file ID format
	if len(fileID) != 32 || !isHexString(fileID) {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	fileInfo, exists := fileStore[fileID]
	if !exists {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"file": %s}`, toJSON(fileInfo))
}

func apiFileInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	// Validate file ID format
	if len(fileID) != 32 || !isHexString(fileID) {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "Invalid file ID"}`, http.StatusBadRequest)
		return
	}

	fileInfo, exists := fileStore[fileID]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "File not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"file": %s}`, toJSON(fileInfo))
}

func validateFile(fileHeader *multipart.FileHeader) error {
	// File size validation
	if fileHeader.Size > maxFileSize {
		return fmt.Errorf("file size too large (max 50MB)")
	}

	// Validate filename to prevent path traversal attacks
	filename := filepath.Base(fileHeader.Filename)
	if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
		return fmt.Errorf("invalid filename")
	}

	// Extension validation
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt", ".zip"}
	allowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("file type not allowed")
	}

	// MIME type validation
	mimeType := fileHeader.Header.Get("Content-Type")
	allowedTypes := []string{
		"image/jpeg", "image/png", "image/gif",
		"application/pdf", "text/plain", "application/zip",
		"application/x-zip-compressed", "application/octet-stream",
	}
	allowed = false
	for _, allowedType := range allowedTypes {
		if mimeType == allowedType {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("MIME type not allowed: %s", mimeType)
	}

	return nil
}

func generateFileID(filename string) string {
	// Generate unique ID using filename and timestamp
	// Sanitize filename to prevent path traversal
	safeFilename := filepath.Base(filename)
	data := fmt.Sprintf("%s-%d", safeFilename, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func toJSON(v interface{}) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// sanitizeFilename removes potentially dangerous characters from filename
func sanitizeFilename(filename string) string {
	// Remove any path separators
	filename = filepath.Base(filename)
	// Remove any control characters and quotes
	filename = strings.Map(func(r rune) rune {
		if r < 32 || r == '"' || r == '\'' {
			return -1
		}
		return r
	}, filename)
	return filename
}
