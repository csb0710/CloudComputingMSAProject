package main

import (
	pb "clcum/protobuf"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lectureDB/common"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

type Course struct {
	TkcrsEcaOrgnm    string `json:"tkcrsEcaOrgnm"`
	Grade            int    `json:"grade"`
	CuriCparNm       string `json:"curiCparNm"`
	SubjId           string `json:"subjId"`
	SubjKnm          string `json:"subjKnm"`
	DvclsNb          int    `json:"dvclsNb"`
	Crd              string `json:"crd"`
	WkLecrEmpnm      string `json:"wkLecrEmpnm"`
	BuldAndRoomCont  string `json:"buldAndRoomCont"`
	CybCoronaTyNm    string `json:"cybCoronaTyNm"`
	enrolledStudents int    `json:"enrolledStudents"`
	maxStudents      int    `json:"maxStudents"`
}

type server struct {
	pb.UnimplementedEnrollmentServiceServer
}

var DB *sql.DB
var ErrMaxCapacityReached = errors.New("cannot increase enrolled students, already at max capacity")

func (s *server) GetCourses(ctx context.Context, req *pb.GetCoursesRequest) (*pb.GetCoursesResponse, error) {
	rows, err := DB.Query("SELECT subjId, tkcrsEcaOrgnm, grade, curiCparNm, subjKnm, dvclsNb, crd, wkLecrEmpnm, buldAndRoomCont, cybCoronaTyNm, enrolledStudents, maxStudents FROM courses")
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var courses []*pb.Course

	for rows.Next() {
		var course pb.Course
		if err := rows.Scan(
			&course.SubjId,
			&course.TkcrsEcaOrgnm,
			&course.Grade,
			&course.CuriCparNm,
			&course.SubjKnm,
			&course.DvclsNb,
			&course.Crd,
			&course.WkLecrEmpnm,
			&course.BuldAndRoomCont,
			&course.CybCoronaTyNm,
			&course.EnrolledStudents,
			&course.MaxStudents,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		courses = append(courses, &course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %v", err)
	}

	return &pb.GetCoursesResponse{
		Courses: courses, // 과목 목록을 포함하는 응답
	}, nil
}

func (s *server) Enroll(ctx context.Context, req *pb.EnrollRequest) (*pb.EnrollResponse, error) {
	// 여기에 데이터베이스 로직 추가
	log.Printf("Received enrollment request: student_id=%s, subj_id=%s\n dvcls_nb_nd=%s", req.StudentId, req.SubjId, req.DvclsNb)

	err := incrementEnrolledStudents(req.SubjId, req.DvclsNb)
	if err != nil {
		if errors.Is(err, ErrMaxCapacityReached) {
			return &pb.EnrollResponse{Success: true, Message: "수강 신청이 마감된 강좌입니다"}, nil
		}
		return &pb.EnrollResponse{Success: false, Message: "Enrollment fail: " + err.Error()}, err
	}

	return &pb.EnrollResponse{Success: true, Message: "수강 신청이 되었습니다"}, nil
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

	file, err := os.Open("courses.json")
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
	var courses []Course
	err = json.Unmarshal(data, &courses)
	if err != nil {
		log.Printf("Failed to parse JSON data: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(1) FROM courses").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check if table is empty: %v", err)
	}

	// If the table is empty, insert the data
	if count == 0 {
		for _, course := range courses {
			insertQuery := fmt.Sprintf("INSERT INTO courses (tkcrsEcaOrgnm, grade, curiCparNm, subjId, subjKnm, dvclsNb, crd, wkLecrEmpnm, buldAndRoomCont, cybCoronaTyNm) VALUES ('%s', %d, '%s', '%s', '%s', %d, '%s', '%s', '%s', '%s')",
				course.TkcrsEcaOrgnm, course.Grade, course.CuriCparNm, course.SubjId, course.SubjKnm, course.DvclsNb, course.Crd, course.WkLecrEmpnm, course.BuldAndRoomCont, course.CybCoronaTyNm)
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
	pb.RegisterEnrollmentServiceServer(s, &server{})

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
	queries := []string{
		`CREATE TABLE IF NOT EXISTS courses (
            subjId VARCHAR(10),
            tkcrsEcaOrgnm VARCHAR(100),
            grade INT,
            curiCparNm VARCHAR(50),
            subjKnm VARCHAR(100),
            dvclsNb VARCHAR(10),
            crd VARCHAR(10),
            wkLecrEmpnm VARCHAR(100),
            buldAndRoomCont VARCHAR(100),
            cybCoronaTyNm VARCHAR(50),
            enrolledStudents INT DEFAULT 0,
            maxStudents INT DEFAULT 40,
			PRIMARY KEY (subjId, dvclsNb)
        );`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// 특정 subjId의 수강신청 인원수 증가
func incrementEnrolledStudents(subjId string, dvclsNb string) error {
	// 트랜잭션 시작
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 현재 enrolledStudents 값을 잠금
	query := `SELECT enrolledStudents, maxStudents FROM courses WHERE subjId = ? and dvclsNb = ? FOR UPDATE`
	var enrolledStudents, maxStudents int
	err = tx.QueryRow(query, subjId, dvclsNb).Scan(&enrolledStudents, &maxStudents)
	if err != nil {
		return err
	}

	// 수강 신청 인원수를 증가시킬 수 있는지 확인
	if enrolledStudents >= maxStudents {
		return fmt.Errorf("%w", ErrMaxCapacityReached)
	}

	// 수강 신청 인원수를 증가
	query = `UPDATE courses SET enrolledStudents = enrolledStudents + 1 WHERE subjId = ? and dvclsNb = ?`
	_, err = tx.Exec(query, subjId, dvclsNb)
	if err != nil {
		return err
	}

	// 트랜잭션 커밋
	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Enrolled students increased successfully")
	return nil
}

// 특정 subjId의 수강신청 인원수 감소
func decrementEnrolledStudents(subjId string) error {
	// 트랜잭션 시작
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 현재 enrolledStudents 값을 잠금
	query := `SELECT enrolledStudents FROM courses WHERE subjId = ? FOR UPDATE`
	var enrolledStudents int
	err = tx.QueryRow(query, subjId).Scan(&enrolledStudents)
	if err != nil {
		return err
	}

	// 수강 신청 인원수를 감소시킬 수 있는지 확인
	if enrolledStudents <= 0 {
		return fmt.Errorf("cannot decrease enrolled students, already at zero")
	}

	// 수강 신청 인원수를 감소
	query = `UPDATE courses SET enrolledStudents = enrolledStudents - 1 WHERE subjId = ?`
	_, err = tx.Exec(query, subjId)
	if err != nil {
		return err
	}

	// 트랜잭션 커밋
	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Println("Enrolled students decreased successfully")
	return nil
}
