// utils/signedurl.go

package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var secretKey = []byte(strings.Trim(os.Getenv("SIGNED_URL_SECRET"), `"`))

func GenerateSignedURL(filename string, baseURL string, ttlSeconds int64) string {
	expires := time.Now().Unix() + ttlSeconds
	toSign := fmt.Sprintf("%s|%d", filename, expires)

	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(toSign))
	signature := hex.EncodeToString(h.Sum(nil))

	u, _ := url.Parse(baseURL)
	u.Path += "/stream/" + filename
	q := u.Query()
	q.Set("expires", strconv.FormatInt(expires, 10))
	q.Set("sig", signature)
	u.RawQuery = q.Encode()

	return u.String()
}

func ValidateSignature(filename string, expiresStr string, sig string) bool {
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expires {
		return false
	}
	toSign := fmt.Sprintf("%s|%d", filename, expires)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(toSign))
	validSig := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(validSig))
}

// Secure stream handler with signature validation
func StreamAudioWithValidation(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	filename := parts[len(parts)-1]
	expires := r.URL.Query().Get("expires")
	sig := r.URL.Query().Get("sig")

	if !ValidateSignature(filename, expires, sig) {
		http.Error(w, "Lien expiré ou signature invalide", http.StatusForbidden)
		return
	}

	path := filepath.Join("audio", filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, path)
}

// Signed URL generator handler (protected route)
func HandleGenerateSignedURL(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "filename manquant", http.StatusBadRequest)
		return
	}
	signedURL := GenerateSignedURL(filename, "http://localhost:8080", 600)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"url":"%s"}`, signedURL)
}
