package httpServe

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

type Handler func(w http.ResponseWriter, r *http.Request)

type Endpoint struct {
	Path    string
	Handler Handler
}

func Listen(port int, endpoints []Endpoint) (*http.Server, error) {
	mux := http.NewServeMux()

	for _, endpoint := range endpoints {
		mux.HandleFunc(endpoint.Path, endpoint.Handler)
	}

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	errorServe := make(chan error, 1)

	go func() {
		log.Println("Starting server on port " + server.Addr)
		err := server.ListenAndServe()

		if err != nil {
			errorServe <- err
		}
	}()

	select {
	case err := <-errorServe:
		log.Printf("Could not start server: %v\n\n", err)
		return nil, err
	default:
	}

	return server, nil
}
