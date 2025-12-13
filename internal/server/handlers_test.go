package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eslutz/forwardarr/internal/qbit"
)

func TestHealthHandler_Running(t *testing.T) {
	server := &Server{
		isRunning: true,
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("healthHandler() status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "OK" {
		t.Errorf("healthHandler() body = %q, want %q", w.Body.String(), "OK")
	}
}

func TestHealthHandler_NotRunning(t *testing.T) {
	server := &Server{
		isRunning: false,
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.healthHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("healthHandler() status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestReadyHandler_Success(t *testing.T) {
	// Create a test HTTP server for qbit
	qbitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ok."))
			return
		}
		if r.URL.Path == "/api/v2/app/version" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("v4.5.0"))
			return
		}
	}))
	defer qbitServer.Close()

	client, _ := qbit.NewClient(qbitServer.URL, "admin", "admin")
	server := &Server{
		qbitClient: client,
	}

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	server.readyHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("readyHandler() status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != "Ready" {
		t.Errorf("readyHandler() body = %q, want %q", w.Body.String(), "Ready")
	}
}

func TestReadyHandler_Failure(t *testing.T) {
	// Create a test HTTP server that always fails for login
	qbitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ok."))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer qbitServer.Close()

	client, _ := qbit.NewClient(qbitServer.URL, "admin", "admin")
	server := &Server{
		qbitClient: client,
	}

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	server.readyHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("readyHandler() status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestStatusHandler_Running(t *testing.T) {
	qbitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ok."))
			return
		}
		if r.URL.Path == "/api/v2/app/version" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("v4.5.0"))
			return
		}
	}))
	defer qbitServer.Close()

	client, _ := qbit.NewClient(qbitServer.URL, "admin", "admin")
	server := &Server{
		qbitClient: client,
		isRunning:  true,
	}

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	server.statusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("statusHandler() status = %d, want %d", w.Code, http.StatusOK)
	}

	var status struct {
		Status               string `json:"status"`
		Version              string `json:"version"`
		QBittorrentReachable bool   `json:"qbittorrent_reachable"`
	}

	err := json.NewDecoder(w.Body).Decode(&status)
	if err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}

	if status.Status != "running" {
		t.Errorf("status.Status = %q, want %q", status.Status, "running")
	}
	if !status.QBittorrentReachable {
		t.Error("status.QBittorrentReachable = false, want true")
	}
}

func TestStatusHandler_Stopping(t *testing.T) {
	qbitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ok."))
			return
		}
	}))
	defer qbitServer.Close()

	client, _ := qbit.NewClient(qbitServer.URL, "admin", "admin")
	server := &Server{
		qbitClient: client,
		isRunning:  false,
	}

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	server.statusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("statusHandler() status = %d, want %d", w.Code, http.StatusOK)
	}

	var status struct {
		Status               string `json:"status"`
		Version              string `json:"version"`
		QBittorrentReachable bool   `json:"qbittorrent_reachable"`
	}

	err := json.NewDecoder(w.Body).Decode(&status)
	if err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}

	if status.Status != "stopping" {
		t.Errorf("status.Status = %q, want %q", status.Status, "stopping")
	}
}

func TestNewServer(t *testing.T) {
	qbitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ok."))
			return
		}
	}))
	defer qbitServer.Close()

	client, _ := qbit.NewClient(qbitServer.URL, "admin", "admin")
	server := NewServer("9090", client)

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	if server.port != "9090" {
		t.Errorf("server.port = %q, want %q", server.port, "9090")
	}
	if !server.isRunning {
		t.Error("server.isRunning = false, want true")
	}
	if server.qbitClient != client {
		t.Error("server.qbitClient not set correctly")
	}
}

func TestSetRunning(t *testing.T) {
	server := &Server{isRunning: true}

	server.SetRunning(false)
	if server.isRunning {
		t.Error("SetRunning(false) did not update isRunning")
	}

	server.SetRunning(true)
	if !server.isRunning {
		t.Error("SetRunning(true) did not update isRunning")
	}
}
