package main

import (
	pb "clcum/protobuf"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"userDB/common"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

type Student struct {
	StudentId string `json:"studentId"`
	Password  string `json:"password"`
}

type server struct {
	pb.UnimplementedStudentServiceServer
}

var DB *sql.DB

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var userExists bool
	query := "SELECT EXISTS(SELECT 1 FROM students WHERE studentId = ? AND password = ?) AS user_exists;"
	err := DB.QueryRow(query, req.StudentId, req.Password).Scan(&userExists)
	if err != nil {
		return &pb.LoginResponse{
			Success: false,
			Message: "Login fail!", // 과목 목록을 포함하는 응답
		}, err
	}

	if userExists {
		return &pb.LoginResponse{
			Success: true,
			Message: "Login Success!", // 과목 목록을 포함하는 응답
		}, nil
	} else {
		return &pb.LoginResponse{
			Success: false,
			Message: "Login fail!", // 과목 목록을 포함하는 응답
		}, nil
	}
}

func main() {
	cfg, err := common.LoadEnvVars()
	if err != nil {
		log.Fatalf("Could not load environment variables: %v", err)
	}
	// MySQL 연결
	dsn := fmt.Sprintf("root:%s@tcp(%s:3306)/%s?charset=utf8mb4", cfg.DBPassword, cfg.DBHost, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	DB = db

	if err := createTables(); err != nil {
		log.Fatalf("Error creating tables: %v", err)
	} else {
		log.Printf("Tables created successfully")
	}

	file, err := os.Open("users.json")
	if err != nil {
		log.Fatalf("Failed to open JSON file: %v", err)
	}
	defer file.Close()

	// 파일 내용 읽기
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	// JSON 데이터 파싱
	var users []Student
	err = json.Unmarshal(data, &users)
	if err != nil {
		log.Printf("Failed to parse JSON data: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(1) FROM students").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check if table is empty: %v", err)
	}

	// If the table is empty, insert the data
	if count == 0 {
		for _, user := range users {
			insertQuery := fmt.Sprintf("INSERT INTO students (studentId, password) VALUES ('%s', '%s')",
				user.StudentId, user.Password)
			_, err := db.Exec(insertQuery)
			if err != nil {
				log.Fatalf("Failed to insert data into database: %v", err)
			}
		}
		log.Println("Data inserted successfully as the table was empty.")
	} else {
		log.Println("Table is not empty, skipping insertion.")
	}

	/////////////
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterStudentServiceServer(s, &server{})

	log.Println("gRPC server listening on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	////////////

	// 신호 처리 설정
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 종료 신호를 받을 때까지 대기
	sig := <-sigChan
	fmt.Printf("Received signal: %v. Shutting down...\n", sig)
}

func createTables() error {
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS students (
        studentId VARCHAR(255) PRIMARY KEY,
        password VARCHAR(255) NOT NULL
    );
    `

	// 테이블 생성
	_, err := DB.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
		return err
	}

	fmt.Println("Table 'students' created successfully!")
	return nil
}
