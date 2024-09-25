package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	pb "project/proto"

	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedAddressServiceServer
	db *sql.DB
}

func (s *server) GetAddressInfo(ctx context.Context, req *pb.AddressRequest) (*pb.AddressResponse, error) {
	placeName := req.GetPlaceName()
	rows, err := s.db.Query(`
	SELECT DISTINCT
        t1.place_name, 
        t1.place_time_start, 
        t1.place_time_end, 
        t1.place_phone, 
        t1.place_email, 
        t1.place_adress, 
        t1.place_latitude, 
        t1.place_longitude 
    FROM 
        adress_contacts.place_info t2
    LEFT JOIN 
        adress_contacts.place t1 
    ON 
        t2.place_info_place_name = t1.place_name
    WHERE 
        t2.place_info_type_place_name = $1
	`, placeName)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []*pb.Place
	for rows.Next() {
		var place pb.Place
		var latitude, longitude float64 // Для хранения широты и долготы
		if err := rows.Scan(&place.PlaceName, &place.PlaceTimeStart, &place.PlaceTimeEnd, &place.PlacePhone, &place.PlaceEmail, &place.PlaceAdress, &latitude, &longitude); err != nil {
			return nil, err
		}

		// Присваиваем значения широты и долготы в структуру Point
		place.PlaceAdressPoint = &pb.Point{
			Latitude:  latitude,
			Longitude: longitude,
		}

		places = append(places, &place)
	}

	return &pb.AddressResponse{Places: places}, nil
}

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=user password=SamaraSamara dbname=db_schedule sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterAddressServiceServer(grpcServer, &server{db: db})

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("Server is running on port 50054...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
