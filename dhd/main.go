package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

// Hard-coded valid address: Crater, Virgo, Sagittarius, Capricornus, Orion, Monoceros + Earth (origin)
// (approximate Abydos coordinates from SG-1, 7 chevrons total)
var validSymbols = []int{1, 2, 11, 14, 29, 31, 39}

var (
	mu       sync.Mutex
	gateOpen bool
	irisOpen bool

	tmpl         *template.Template
	externalBase = getEnv("EXTERNAL_API", "http://localhost:9000")
)

func main() {
	var err error
	tmpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", page("index.html"))
	mux.HandleFunc("GET /dhd", page("dhd.html"))
	mux.HandleFunc("GET /scdc", page("scdc.html"))
	mux.HandleFunc("POST /api/dial", handleDial)
	mux.HandleFunc("POST /api/disconnect", handleDisconnect)
	mux.HandleFunc("POST /api/iris", handleIris)
	mux.HandleFunc("GET /api/state", handleState)
	mux.Handle("GET /sounds/", http.StripPrefix("/sounds/", http.FileServer(http.Dir("sounds"))))

	addr := ":" + getEnv("PORT", "8080")
	log.Printf("stargate control listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func page(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.ExecuteTemplate(w, name, nil); err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}

// ── request / response types ──────────────────────────────────────────────────

type DialReq struct {
	Symbols []int `json:"symbols"`
}

type Resp struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	State   State  `json:"state"`
}

type State struct {
	GateOpen bool `json:"gateOpen"`
	IrisOpen bool `json:"irisOpen"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

func handleDial(w http.ResponseWriter, r *http.Request) {
	var req DialReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, Resp{OK: false, Message: "invalid request"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if gateOpen {
		writeJSON(w, 200, Resp{OK: false, Message: "wormhole already active", State: snap()})
		return
	}
	if !matchAddr(req.Symbols, validSymbols) {
		writeJSON(w, 200, Resp{OK: false, Message: "no gate found at those coordinates", State: snap()})
		return
	}
	if err := callExt("POST", "/dial", nil); err != nil {
		log.Printf("external /dial: %v", err)
		writeJSON(w, 502, Resp{OK: false, Message: "gate malfunction — no response from hardware"})
		return
	}
	gateOpen = true
	writeJSON(w, 200, Resp{OK: true, Message: "wormhole established", State: snap()})
}

func handleDisconnect(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if !gateOpen {
		writeJSON(w, 200, Resp{OK: false, Message: "no active wormhole", State: snap()})
		return
	}
	if err := callExt("POST", "/disconnect", nil); err != nil {
		log.Printf("external /disconnect: %v", err)
		writeJSON(w, 502, Resp{OK: false, Message: "disconnect failed"})
		return
	}
	gateOpen = false
	writeJSON(w, 200, Resp{OK: true, Message: "wormhole disengaged", State: snap()})
}

func handleIris(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	action := "open"
	if irisOpen {
		action = "close"
	}
	if err := callExt("POST", "/iris/"+action, nil); err != nil {
		log.Printf("external /iris: %v", err)
		writeJSON(w, 502, Resp{OK: false, Message: "iris malfunction"})
		return
	}
	irisOpen = !irisOpen
	writeJSON(w, 200, Resp{OK: true, Message: "iris " + action + "d", State: snap()})
}

func handleState(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, 200, Resp{OK: true, State: snap()})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func snap() State { return State{GateOpen: gateOpen, IrisOpen: irisOpen} }

func matchAddr(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	a, b := make([]int, len(got)), make([]int, len(want))
	copy(a, got)
	copy(b, want)
	sort.Ints(a)
	sort.Ints(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func callExt(method, path string, body any) error {
	var br io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		br = bytes.NewReader(data)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, externalBase+path, br)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
