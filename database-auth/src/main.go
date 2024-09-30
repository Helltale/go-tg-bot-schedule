package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	pb "project/proto"

	_ "github.com/lib/pq" // PostgreSQL driver
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedAuthServer
	db *sql.DB
}

func (s *server) CheckUser(ctx context.Context, req *pb.CheckUserRequest) (*pb.CheckUserResponse, error) {
	var profileName, roleName string
	err := s.db.QueryRow("SELECT profile_name, profile_role_name FROM account.profile WHERE profile_tg_id = $1", req.GetProfileTgId()).Scan(&profileName, &roleName)

	if err == sql.ErrNoRows {
		// Пользователь не найден
		log.Printf("пользователь `%d` не найден в системе, запуск регистрации", req.GetProfileTgId())
		return &pb.CheckUserResponse{Exists: false}, nil
	} else if err != nil {
		return nil, fmt.Errorf("ошибка при проверке пользователя: %v", err)
	}

	log.Printf("пользователь `%d` найден в системе, rolename `%s`", req.GetProfileTgId(), roleName)
	// Пользователь найден
	return &pb.CheckUserResponse{Exists: true, ProfileName: profileName, RoleName: roleName}, nil
}

func (s *server) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	// Вставляем пользователя в таблицу account.profile
	_, err := s.db.Exec("INSERT INTO account.profile (profile_tg_id, profile_name, profile_role_name) VALUES ($1, $2, $3)",
		req.GetProfileTgId(), req.GetProfileName(), "Студент")
	if err != nil {
		return nil, fmt.Errorf("ошибка при регистрации пользователя: %v", err)
	}
	log.Printf("зареган пользователь `%d` `%s` `%s` \n", req.GetProfileTgId(), req.GetProfileName(), "Студент")

	// Вставляем данные о группе в таблицу schedule.group
	_, err = s.db.Exec("INSERT INTO schedule.group (group_type_group_name, group_profile_tg_id) VALUES ($1, $2)",
		req.GetGroupName(), req.GetProfileTgId())
	if err != nil {
		return nil, fmt.Errorf("ошибка при добавлении пользователя в группу: %v", err)
	}
	log.Printf("пользователь `%d` занесен в группу `%s`", req.GetProfileTgId(), req.GetGroupName())

	return &pb.RegisterUserResponse{Success: true}, nil
}

func (s *server) GetGroups(ctx context.Context, req *pb.Empty) (*pb.GetGroupsResponse, error) {
	// Создаем срез для хранения названий групп
	var groups []string

	// Выполняем SQL-запрос для получения всех названий групп
	rows, err := s.db.Query("SELECT type_group_name FROM schedule.type_group")
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса к базе данных: %v", err)
	}
	defer rows.Close()

	// Проходим по результатам запроса
	for rows.Next() {
		var groupName string
		if err := rows.Scan(&groupName); err != nil {
			return nil, fmt.Errorf("ошибка при считывании названия группы: %v", err)
		}
		groups = append(groups, groupName) // Добавляем название группы в срез
	}

	// Проверяем на наличие ошибок после завершения итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов запроса: %v", err)
	}

	// Возвращаем ответ с названиями групп
	return &pb.GetGroupsResponse{Groups: groups}, nil
}

func main() {
	// Подключение к базе данных
	db, err := sql.Open("postgres", "host=localhost port=5432 user=user password=SamaraSamara dbname=db_schedule sslmode=disable")
	if err != nil {
		log.Fatalf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	s := grpc.NewServer()
	pb.RegisterAuthServer(s, &server{db: db})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("ошибка при прослушивании: %v", err)
	}

	log.Println("Сервер АВТОРИЗАЦИИ запущен на порту :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("ошибка при запуске сервера: %v", err)
	}
}
