package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"userBE/common"

	pb "clcum/protobuf"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	// 프로토버퍼로 생성된 패키지 경로로 수정
)

type LoginRequest struct {
	StudentID string `json:"studentId"`
	Password  string `json:"password"`
}

var loginClient pb.StudentServiceClient
var enrollCTX context.Context

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loginReq := &pb.LoginRequest{
		StudentId: req.StudentID,
		Password:  req.Password,
	}
	res, err := loginClient.Login(enrollCTX, loginReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/login", loginHandler).Methods("POST")

	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	fmt.Println("Starting server on :8080")
	go func() {
		log.Fatal(http.ListenAndServe(":8080", corsMiddleware(r)))
	}()

	cfg, err := common.LoadEnvVars()
	if err != nil {
		log.Fatalf("Could not load environment variables: %v", err)
	}

	// Construct address and start listening
	addr := fmt.Sprintf("%s:%d", cfg.ServerAddr, cfg.ServerPort)

	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	// Start serving gRPC server
	log.Printf("[gRPC] Successfully connected to %s", addr)

	client := pb.NewStudentServiceClient(conn)
	loginClient = client

	// Enroll 요청 보내기
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx := context.Background()
	// defer cancel()
	enrollCTX = ctx

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
}
