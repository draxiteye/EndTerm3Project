package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
    "os"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type SystemStatus string

const (
	StatusHealthy  SystemStatus = "HEALTHY"
	StatusWarning  SystemStatus = "WARNING"
	StatusCritical SystemStatus = "CRITICAL"
	StatusDead     SystemStatus = "DEAD"
)

type ServerData struct {
	ID        string       `json:"id"`
	Status    SystemStatus `json:"status"`
	CPU       float64      `json:"cpu"`
	Memory    float64      `json:"memory"`
	Latency   float64      `json:"latency"`
	Disk      float64      `json:"disk"`
	ReqRate   int          `json:"reqRate"`
	ErrorRate float64      `json:"errorRate"`
	Chaos     string       `json:"chaos,omitempty"`
}

type Simulation struct {
	mu      sync.Mutex
	Servers []*ServerData
	Logs    []string
}

func NewSimulation(count int) *Simulation {
	sim := &Simulation{
		Servers: make([]*ServerData, count),
		Logs:    []string{"[SYSTEM] Simulation Engine Initialized."},
	}
	for i := 0; i < count; i++ {
		sim.Servers[i] = &ServerData{
			ID:     fmt.Sprintf("node-%03d", i+1),
			Status: StatusHealthy,
			Disk:   rand.Float64() * 30,
		}
	}
	return sim
}

func (s *Simulation) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, srv := range s.Servers {
		if srv.Status == StatusDead {
			if rand.Float64() < 0.03 {
				srv.Status = StatusHealthy
				s.Logs = append(s.Logs, fmt.Sprintf("[%s] %s recovered.", time.Now().Format("15:04:05"), srv.ID))
			}
			continue
		}

		// Fluctuations
		srv.CPU = clamp(srv.CPU+(rand.Float64()*12-6), 2, 100)
		srv.Memory = clamp(srv.Memory+(rand.Float64()*6-3), 5, 100)
		srv.Latency = clamp(srv.Latency+(rand.Float64()*40-20), 5, 800)
		srv.ErrorRate = clamp(srv.ErrorRate+(rand.Float64()*4-2), 0, 100)
		srv.ReqRate = int(clamp(float64(srv.ReqRate)+(rand.Float64()*80-40), 0, 2000))

		// Random Chaos
		if rand.Float64() < 0.01 {
			srv.Status = StatusCritical
			srv.Chaos = "NETWORK_JITTER"
			if rand.Float64() < 0.3 {
				srv.Status = StatusDead
				s.Logs = append(s.Logs, fmt.Sprintf("[%s] %s connection lost.", time.Now().Format("15:04:05"), srv.ID))
			}
		} else {
			srv.Chaos = ""
		}

		// Status logic
		if srv.Status != StatusDead {
			if srv.CPU > 85 || srv.ErrorRate > 40 {
				srv.Status = StatusCritical
			} else if srv.CPU > 60 || srv.Latency > 300 {
				srv.Status = StatusWarning
			} else {
				srv.Status = StatusHealthy
			}
		}
	}
	if len(s.Logs) > 40 { s.Logs = s.Logs[1:] }
}

func clamp(val, min, max float64) float64 {
	if val < min { return min }; if val > max { return max }; return val
}

func main() {
	sim := NewSimulation(100)

	go func() {
		for {
			sim.Tick()
			time.Sleep(800 * time.Millisecond)
		}
	}()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		for {
			sim.mu.Lock()
			payload := map[string]interface{}{"servers": sim.Servers, "logs": sim.Logs}
			sim.mu.Unlock()
			if err := conn.WriteJSON(payload); err != nil { break }
			time.Sleep(800 * time.Millisecond)
		}
	})

	// Add a chaos trigger endpoint
	http.HandleFunc("/chaos", func(w http.ResponseWriter, r *http.Request) {
		sim.mu.Lock()
		for i := 0; i < 10; i++ {
			idx := rand.Intn(len(sim.Servers))
			sim.Servers[idx].Status = StatusDead
		}
		sim.Logs = append(sim.Logs, "[ALARM] Manual Chaos Event Triggered!")
		sim.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("🚀 Pro-Observability Backend on :8080")
	port := os.Getenv("PORT")
if port == "" {
	port = "8080"
}

fmt.Println("🚀 Server running on port", port)
log.Fatal(http.ListenAndServe(":"+port, nil))
}