package rest

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/auth"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/engine"
	"github.com/theshamuel/image-jobs-dispatcher/dispatcher/app/model"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Rest struct {
	Version          string
	WorkerServiceURI string
	RemoteService    engine.Interface
	httpServer       *http.Server
	Auth             *auth.Service
	lock             sync.Mutex
}

type request struct {
	Encoding string `json:"encoding"`
	MD5      string `json:"md5"`
	Data     string `json:"content"`
}

const sizeBodyLimit = 1024 * 1024 * 3 // limit size of request body

//Run http server
func (r *Rest) Run(port int) {
	log.Printf("[INFO] Run http server on port %d", port)
	r.lock.Lock()
	r.httpServer = r.buildHTTPServer(port, r.routes())
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

//Default body size 10Mb from request.go
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
			api.Get("/job/{id}", r.getJob)
		})
	})

	return router
}

func (r *Rest) getJobStatus(w http.ResponseWriter, req *http.Request) {
	jobID := chi.URLParam(req, "id")
	_, err := r.checkJWT(req.Header.Get("Authorization"), &w)
	if err != nil {
		SendErrorJSON(w, req, http.StatusUnauthorized, err, ErrorMD5Validation, "JWT is invalid")
		return
	}
	status, err := r.RemoteService.GetStatusJob(jobID)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during getting job status in worker service")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte("{ \"status\":\"" + status.ToString() + "\"}"))
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during writing response")
		return
	}
}

func (r *Rest) submitJob(w http.ResponseWriter, req *http.Request) {
	msg := request{}
	if err := render.DecodeJSON(http.MaxBytesReader(w, req.Body, sizeBodyLimit), &msg); err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorJSONUnmarshal, "can't unmarshal request message")
		return
	}
	if err := msg.checkMd5Hash(); err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorMD5Validation, "Error during md5 validation")
		return
	}
	claims, err := r.checkJWT(req.Header.Get("Authorization"), &w)
	if err != nil {
		SendErrorJSON(w, req, http.StatusUnauthorized, err, ErrorMD5Validation, "JWT is invalid")
		return
	}
	job := model.Job{ClientID: claims.ClientID,
		TenantID:    claims.TenantID,
		Payload:     msg.Data,
		PayloadSize: len(msg.Data)}

	resJob, err := r.RemoteService.SubmitJob(job)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during submitting job in worker service")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(resJob)
	if err != nil {
		log.Printf("[ERROR] can not encode response body %#v", err)
		SendErrorJSON(w, req, http.StatusUnauthorized, err, ErrorMD5Validation, "error during encoding job response")
		return
		return
	}
	if _, err = w.Write(body); err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during writing response")
		return
	}
}

func (r *Rest) checkJWT(authHeader string, w *http.ResponseWriter) (auth.Claims, error) {
	if authHeader == "" || len(strings.Split(authHeader, " ")) != 2 {
		(*w).WriteHeader(http.StatusUnauthorized)
		return auth.Claims{}, nil
	}
	headerValue := strings.Split(authHeader, " ")
	return r.Auth.Parse(headerValue[1])
}

func (msg *request) checkMd5Hash() error {
	hash := md5.New()
	decodedData, err := decodeByAlgorithm([]byte(msg.Data), msg.Encoding)
	if err != nil {
		log.Printf("[ERROR] can't decode base64: %s", err)
		return err
	}

	if _, err = io.Copy(hash, bytes.NewReader(decodedData)); err != nil {
		log.Printf("[ERROR] can't copy image byte in creating Reader image: %s", err)
		return err
	}

	if string(hash.Sum(nil)) == msg.MD5 {
		return fmt.Errorf("MD5 hash sum is not valid passed: %s, calculated: %x", msg.MD5, hash.Sum(nil))
	}

	return nil
}

func decodeByAlgorithm(d []byte, alg string) ([]byte, error) {
	switch alg {
	case "base64":
		return base64Decode(d)
	default:
		return nil, errors.New("unknown encoding algorithm")
	}
}

func base64Decode(data []byte) ([]byte, error) {
	var size int
	res := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	size, err := base64.StdEncoding.Decode(res, data)
	if err != nil {
		return nil, err
	}
	return res[:size], nil
}

func (r *Rest) getJob(w http.ResponseWriter, req *http.Request) {
	jobID := chi.URLParam(req, "id")
	_, err := r.checkJWT(req.Header.Get("Authorization"), &w)
	if err != nil {
		SendErrorJSON(w, req, http.StatusUnauthorized, err, ErrorMD5Validation, "JWT is invalid")
		return
	}
	job, err := r.RemoteService.GetJob(jobID)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during getting job")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(job)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during marshal response")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err = w.Write(data); err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during writing response")
		return
	}
}
