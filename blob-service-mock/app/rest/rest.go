package rest

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Rest struct {
	httpServer *http.Server
	lock       sync.Mutex
}

type ImgID struct {
	ID int `json:"img_id"`
}

//Run http server
func (r *Rest) Run() {
	log.Printf("[INFO] run http server on port %d", 8081)
	r.lock.Lock()
	r.httpServer = r.buildHTTPServer(8081, r.routes())
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
			api.Post("/blob", r.submitBlob)
			api.Get("/blob/{id}", r.getBlob)
		})
	})

	return router
}

func (r *Rest) submitBlob(w http.ResponseWriter, req *http.Request) {
	jsr := ImgID{ID: 1}
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(jsr)
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error during marshal response")
		return
	}
	if _, err := w.Write(data); err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error saving image to blob store")
		return
	}
}

func (r *Rest) getBlob(w http.ResponseWriter, req *http.Request) {
	blobID, err := strconv.Atoi(chi.URLParam(req, "id"))
	if err != nil {
		SendErrorJSON(w, req, http.StatusBadRequest, errors.New("store is overloaded"), ErrorServerInternal, "error getting image to id is not int32")
		return
	}
	if blobID == 1 || blobID == 2 || blobID == 3 {
		w.Header().Set("Content-Type", "image/png")
		f, err := os.Open(fmt.Sprintf("/data/%d", blobID))
		if err != nil {
			SendErrorJSON(w, req, http.StatusBadRequest, errors.New("store is overloaded"), ErrorServerInternal, "error open image from file system")
			return
		}
		reader := bufio.NewReader(f)
		content, err := ioutil.ReadAll(reader)
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		if err != nil {
			SendErrorJSON(w, req, http.StatusBadRequest, errors.New("store is overloaded"), ErrorServerInternal, "error read image from file system")
			return
		}
		if _, err := w.Write(content); err != nil {
			SendErrorJSON(w, req, http.StatusBadRequest, err, ErrorServerInternal, "error saving image to blob store")
			return
		}
	}else {
		SendErrorJSON(w, req, http.StatusBadRequest, errors.New("store is overloaded"), ErrorServerInternal, "error getting image to blob store")
		return
	}
}

