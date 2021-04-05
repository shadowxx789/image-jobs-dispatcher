package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Rest struct {
	httpServer *http.Server
	lock       sync.Mutex
}

type Job struct {
	ID              string    `json:"id"`
	TenantID        int       `json:"tenant_id"`
	ClientID        int       `json:"client_id"`
	Payload         string    `json:"payload,omitempty"`
	PayloadLocation string    `json:"payload_location,omitempty"`
	PayloadSize     int       `json:"payload_size,omitempty"`
	Status          JobStatus `json:"status,omitempty"`
}

type JobStatus int

const (
	RUNNING = iota
	SUCCESS
	FAILED
)

func (j JobStatus) String() string {
	switch j {
	case RUNNING:
		return "RUNNING"
	case SUCCESS:
		return "SUCCESS"
	case FAILED:
		return "FAILED"
	default:
		return fmt.Sprintf("%d", int(j))
	}
}

type JobStatusResponse struct {
	Status int `json:"status"`
}

type ImgID struct {
	PayloadLocation string `json:"payload_location"`
}

var store = map[string]Job{
	"1": {ID: "1", TenantID: 1, ClientID: 1, Status: SUCCESS},
	"2": {ID: "2", TenantID: 2, ClientID: 2, Status: RUNNING},
	"3": {ID: "3", TenantID: 3, ClientID: 3, Status: FAILED},
}

//Run http server
func (r *Rest) Run() {
	log.Printf("[INFO] run http server on port %d", 8080)
	r.lock.Lock()
	r.httpServer = r.buildHTTPServer(8080, r.routes())
	r.lock.Unlock()
	err := r.httpServer.ListenAndServe()
	log.Printf("[WARN] http server terminated, %s", err)
}

// Shutdown http server
func (r *Rest) Shutdown() {
	log.Println("[WARN] shutdown http server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r.lock.Lock()
	if r.httpServer != nil {
		if err := r.httpServer.Shutdown(ctx); err != nil {
			log.Printf("[ERROR] http shutdown error, %s", err)
		}
		log.Println("[DEBUG] shutdown http server completed")
	}
	r.lock.Unlock()
}

func (r *Rest) buildHTTPServer(port int, router http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}

func (r *Rest) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Throttle(1000), middleware.RealIP, middleware.Recoverer, middleware.Logger)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Content-Length", "X-XSRF-Token"},
		ExposedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	//health check api
	router.Use(corsMiddleware.Handler)
	router.Route("/", func(api chi.Router) {
		api.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(5, nil)))
		api.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(fmt.Sprintln("pong")))
			if err != nil {
				log.Printf("[ERROR] cannot write response #%v", err)
			}
		})
	})

	router.Route("/api/v1/", func(endpoints chi.Router) {
		endpoints.Group(func(api chi.Router) {
			api.Use(middleware.Timeout(30 * time.Second))
			api.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(50, nil)))
			api.Use(middleware.NoCache)
			api.Post("/job", r.submitJob)
			api.Get("/job/{id}/status", r.getJobStatus)
		})
	})

	return router
}

func (r *Rest) getJobStatus(w http.ResponseWriter, req *http.Request) {
	jobID := chi.URLParam(req, "id")
	if jobID != "1" && jobID != "2" && jobID != "3" {
		SendErrorJSON(w, req, http.StatusNotFound, fmt.Errorf("there is no job with id %s", jobID), ErrorServerInternal, "error getting job status")
		return
	}
	if _, ok := store[jobID]; !ok {
		SendErrorJSON(w, req, http.StatusNotFound, fmt.Errorf("there is no job with id %s", jobID), ErrorServerInternal, "error getting job status")
		return
	}
	jsr := JobStatusResponse{Status: int(store[jobID].Status)}
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(jsr)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during marshal response")
		return
	}
	_, err = w.Write(data)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during writing response")
		return
	}
}

func (r *Rest) submitJob(w http.ResponseWriter, req *http.Request) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "cannot read request body")
		log.Printf("[ERROR] cannot create POST request")
		return
	}

	request, err := http.NewRequest("POST", "http://localhost:8081/api/v1/blob", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "image/png")
	request.Header.Set("Content-Length", strconv.Itoa(len(body)))
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "cannot create POST request")
		log.Printf("[ERROR] cannot create POST request")
		return
	}

	response, err := client.Do(request)
	if err != nil || response.StatusCode != http.StatusOK {
		log.Printf("[ERROR] can not make post request: %#v", err)
		if errClose := response.Body.Close(); errClose != nil {
			log.Printf("[ERROR] can not close response body %#v", errClose)
		}
		SendErrorJSON(w, req, response.StatusCode, err, ErrorServerInternal, "error during request to blob service")
		return
	}
	if response == nil && err != nil {
		if errClose := response.Body.Close(); errClose != nil {
			log.Printf("[ERROR] can not close response body %#v", errClose)
		}
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "cannot read response from blob service")
		return
	}
	defer response.Body.Close()
	payloadLocation := &ImgID{}
	if err = json.NewDecoder(response.Body).Decode(payloadLocation); err != nil {
		log.Printf("[ERROR] can not decode response body %#v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(payloadLocation)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during marshal response")
		return
	}
	_, err = w.Write(data)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during writing response")
		return
	}
}
