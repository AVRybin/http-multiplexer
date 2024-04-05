package main

import (
	"context"
	"encoding/json"
	"fmt"
	"http-multiplexer/lib/envReader"
	"http-multiplexer/lib/httpClientMultiplexer"
	"http-multiplexer/lib/httpServe"
	"http-multiplexer/schemas"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	Config       schemas.Config
	countRequest int32
)

func main() {
	atomic.StoreInt32(&countRequest, 0)
	err := envReader.GetEnvVariable(&Config)

	if err != nil {
		log.Fatal(err)
	}

	endpoint := httpServe.Endpoint{
		Path:    Config.ServerPath,
		Handler: HandlerHttp,
	}

	server, err := httpServe.Listen(Config.ServerPort, []httpServe.Endpoint{endpoint})

	if err != nil {
		log.Fatal(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exiting")

}

func HandlerHttp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	numberReq := atomic.LoadInt32(&countRequest)

	if int(numberReq) >= Config.ServerMaxCountReq {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Too much request")
		return
	}

	atomic.AddInt32(&countRequest, 1)

	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var Req schemas.RequestBody
	err = json.Unmarshal(reqBody, &Req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	CountUrl := len(Req.UrlList)

	if CountUrl > Config.HttpMaxCountURL {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Max url %v", Config.HttpMaxCountURL)
		return
	}

	var (
		result schemas.ResponseGeneral
	)

	result.CountRequest = CountUrl

	Responses, err := httpClientMultiplexer.SendMultiplexerRequests(r.Context(), Req.UrlList, "GET",
		time.Duration(Config.HttpTimeout)*time.Millisecond, Config.HttpMaxParallelReq)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error url request %v", err)
		log.Println(err)
		return
	}

	result.Responses = Responses

	response, err := json.MarshalIndent(result, "", "\t")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error body decode")
		return
	}

	atomic.AddInt32(&countRequest, -1)
	fmt.Fprintf(w, string(response))
}
