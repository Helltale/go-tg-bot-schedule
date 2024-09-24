package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	pb "project/proto" // Замените на путь к вашему сгенерированному пакету

	_ "github.com/lib/pq" // Импортируйте драйвер базы данных, например, PostgreSQL
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTeacherServiceServer
	db *sql.DB
}

func (s *server) FindTeachersByFIO(ctx context.Context, req *pb.FindTeachersRequest) (*pb.FindTeachersResponse, error) {
	fio := req.Fio
	query := `
        SELECT t.teacher_profile_tg_id, t.teacher_name, t.teacher_job, t.teacher_department, 
               t.teacher_adress, t.teacher_email, i.name_img
        FROM teacher_info.teacher t
        LEFT JOIN teacher_info.image i ON t.teacher_profile_tg_id = i.teacher_id
        WHERE LOWER(t.teacher_name) LIKE LOWER($1)`
	rows, err := s.db.Query(query, "%"+fio+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []*pb.Teacher
	for rows.Next() {
		var teacher pb.Teacher
		if err := rows.Scan(&teacher.TeacherProfileTgId, &teacher.TeacherName, &teacher.TeacherJob,
			&teacher.TeacherDepartment, &teacher.TeacherAdress, &teacher.TeacherEmail,
			&teacher.ImageName); err != nil {
			return nil, err
		}
		teachers = append(teachers, &teacher)
	}

	return &pb.FindTeachersResponse{Teachers: teachers}, nil
}

func (s *server) FindTeachersByDepartment(ctx context.Context, req *pb.FindTeachersByDepartmentRequest) (*pb.FindTeachersResponse, error) {
	department := req.Department
	query := `
        SELECT t.teacher_profile_tg_id, t.teacher_name, t.teacher_job, t.teacher_department, 
               t.teacher_adress, t.teacher_email, i.name_img
        FROM teacher_info.teacher t
        LEFT JOIN teacher_info.image i ON t.teacher_profile_tg_id = i.teacher_id
        WHERE LOWER(t.teacher_department) LIKE LOWER($1)`
	rows, err := s.db.Query(query, "%"+department+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []*pb.Teacher
	for rows.Next() {
		var teacher pb.Teacher
		if err := rows.Scan(&teacher.TeacherProfileTgId, &teacher.TeacherName, &teacher.TeacherJob,
			&teacher.TeacherDepartment, &teacher.TeacherAdress, &teacher.TeacherEmail,
			&teacher.ImageName); err != nil {
			return nil, err
		}
		teachers = append(teachers, &teacher)
	}

	return &pb.FindTeachersResponse{Teachers: teachers}, nil
}

func (s *server) FindTeachersBySubject(ctx context.Context, req *pb.FindTeachersBySubjectRequest) (*pb.FindTeachersResponse, error) {
	subject := req.Subject
	query := `
		SELECT 
			t.teacher_profile_tg_id, 
			t.teacher_name, 
			t.teacher_job, 
			t.teacher_department, 
			t.teacher_adress, 
			t.teacher_email, 
			i.name_img
		FROM teacher_info.teacher t
		LEFT JOIN teacher_info.image i ON t.teacher_profile_tg_id = i.teacher_id
		WHERE t.teacher_profile_tg_id IN (
			SELECT DISTINCT lesson_teacher_id
			FROM schedule.lesson
			WHERE LOWER(lesson_subject_name) LIKE LOWER($1)
		)
	`
	rows, err := s.db.Query(query, "%"+subject+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []*pb.Teacher
	for rows.Next() {
		var teacher pb.Teacher
		if err := rows.Scan(&teacher.TeacherProfileTgId, &teacher.TeacherName, &teacher.TeacherJob,
			&teacher.TeacherDepartment, &teacher.TeacherAdress, &teacher.TeacherEmail,
			&teacher.ImageName); err != nil {
			return nil, err
		}
		teachers = append(teachers, &teacher)
	}

	return &pb.FindTeachersResponse{Teachers: teachers}, nil
}

func main() {
	// Подключение к базе данных
	db, err := sql.Open("postgres", "host=localhost port=5432 user=user password=SamaraSamara dbname=db_schedule sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Ошибка при прослушивании: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTeacherServiceServer(grpcServer, &server{db: db})

	log.Println("Сервер запущен на порту :50053")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
