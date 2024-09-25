package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	pb "project/proto" // Замените на путь к вашему сгенерированному пакету

	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedDocumentServiceServer
	db *sql.DB
}

func (s *server) GetDocuments(ctx context.Context, req *pb.DocumentRequest) (*pb.DocumentListResponse, error) {
	var documents []*pb.DocumentResponse

	// Изменяем запрос, чтобы он использовал только file_group
	query := `
		SELECT d.file_name
		FROM documents.document d
		WHERE d.file_group = $1
	`
	rows, err := s.db.Query(query, req.FileGroup)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var document pb.DocumentResponse
		if err := rows.Scan(&document.FileName); err != nil {
			return nil, err
		}
		documents = append(documents, &document)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Возвращаем список названий файлов
	return &pb.DocumentListResponse{Documents: documents}, nil
}

func main() {
	// Подключение к базе данных
	db, err := sql.Open("postgres", "host=localhost port=5432 user=user password=SamaraSamara dbname=db_schedule sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s := grpc.NewServer()
	pb.RegisterDocumentServiceServer(s, &server{db: db})

	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Сервер запущен на порту :50055")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
