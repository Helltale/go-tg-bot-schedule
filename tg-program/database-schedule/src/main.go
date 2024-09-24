package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	pb "project/proto"

	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedScheduleServiceServer
	db *sql.DB
}

func (s *server) GetSchedule(ctx context.Context, req *pb.ScheduleRequest) (*pb.ScheduleResponse, error) {
	var query string
	if req.RequestType == "today" {
		query = `SELECT lesson_type_education_name, profile_name, lesson_subject_name, lesson_date_time_start, lesson_date_time_end, lesson_link
                  FROM schedule.lesson
                  JOIN account.profile ON lesson_teacher_id = profile_tg_id
                  WHERE lesson_date_time_start::date = CURRENT_DATE AND lesson_type_group_name = $1
				  ORDER BY lesson_date_time_start`
		log.Printf("selected schedule today. for `%s` group", req.GroupName)
	} else if req.RequestType == "week" {
		query = `SELECT lesson_type_education_name, profile_name, lesson_subject_name, lesson_date_time_start, lesson_date_time_end, lesson_link
                  FROM schedule.lesson
                  JOIN account.profile ON lesson_teacher_id = profile_tg_id
                  WHERE lesson_date_time_start >= CURRENT_DATE AND lesson_date_time_start < CURRENT_DATE + INTERVAL '7 days' AND lesson_type_group_name = $1
				  ORDER BY lesson_date_time_start`
		log.Printf("selected schedule week. for `%s` group", req.GroupName)
	} else {
		log.Printf("bad request for `%s` group", req.GroupName)
		return nil, fmt.Errorf("неизвестный тип запроса: %s", req.RequestType)
	}

	rows, err := s.db.Query(query, req.GroupName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []*pb.Lesson
	for rows.Next() {
		var lesson pb.Lesson
		var lessonLink sql.NullString // Используем sql.NullString для обработки NULL значений
		if err := rows.Scan(&lesson.TypeEducation, &lesson.TeacherName, &lesson.SubjectName, &lesson.StartTime, &lesson.EndTime, &lessonLink); err != nil {
			return nil, err
		}
		if lessonLink.Valid {
			lesson.Link = lessonLink.String // Если значение не NULL, присваиваем его
		} else {
			lesson.Link = "" // Или присваиваем пустую строку, если значение NULL
		}
		lessons = append(lessons, &lesson)
	}
	log.Println("Результат:", lessons)
	return &pb.ScheduleResponse{Lessons: lessons}, nil
}

func (s *server) GetGroupByTGID(ctx context.Context, req *pb.GetGroupByTGIDRequest) (*pb.GetGroupByTGIDResponse, error) {
	var groupName string
	query := `SELECT group_type_group_name FROM schedule.group WHERE group_profile_tg_id = $1`
	err := s.db.QueryRow(query, req.ProfileTgId).Scan(&groupName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("группа не найдена для tg_id: %d", req.ProfileTgId)
		}
		return nil, err
	}
	return &pb.GetGroupByTGIDResponse{GroupName: groupName}, nil
}

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=user password=SamaraSamara dbname=db_schedule sslmode=disable")
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterScheduleServiceServer(grpcServer, &server{db: db})

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Ошибка при прослушивании: %v", err)
	}

	log.Println("Сервер с РАСПИСАНИЕМ запущен на порту :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
