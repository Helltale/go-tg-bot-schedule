package clients

import (
	"context"
	"fmt"
	pbac "tgclient/proto/adress-contact"
	pba "tgclient/proto/auth"
	pbs "tgclient/proto/schedule"
	pbt "tgclient/proto/teacher"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetGroupByTGID(ctx context.Context, tgID int64) (string, error) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbs.NewScheduleServiceClient(conn)
	resp, err := client.GetGroupByTGID(ctx, &pbs.GetGroupByTGIDRequest{ProfileTgId: tgID})
	if err != nil {
		return "", fmt.Errorf("ошибка при вызове gRPC метода GetGroupByTGID: %v", err)
	}

	return resp.GroupName, nil
}

func GetSchedule(ctx context.Context, groupName, requestType string) (*pbs.ScheduleResponse, error) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbs.NewScheduleServiceClient(conn)
	resp, err := client.GetSchedule(ctx, &pbs.ScheduleRequest{
		GroupName:   groupName,
		RequestType: requestType,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода GetSchedule: %v", err)
	}

	return resp, nil
}

func CheckUser(ctx context.Context, userID int64) (*pba.CheckUserResponse, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pba.NewAuthClient(conn)
	resp, err := client.CheckUser(ctx, &pba.CheckUserRequest{ProfileTgId: userID})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода CheckUser: %v", err)
	}

	return resp, nil
}

func RegisterUser(ctx context.Context, userID int64, profileName, groupName string) (*pba.RegisterUserResponse, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pba.NewAuthClient(conn)
	resp, err := client.RegisterUser(ctx, &pba.RegisterUserRequest{
		ProfileTgId: userID,
		ProfileName: profileName,
		GroupName:   groupName, // Передаем группу
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода RegisterUser: %v", err)
	}

	return resp, nil
}

func GetGroups(ctx context.Context) (*pba.GetGroupsResponse, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pba.NewAuthClient(conn)
	resp, err := client.GetGroups(ctx, &pba.Empty{})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода GetGroups: %v", err)
	}

	return resp, nil
}

// Функция для поиска преподавателей по ФИО
func FindTeachersByFIO(ctx context.Context, fio string) ([]*pbt.Teacher, error) {
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbt.NewTeacherServiceClient(conn)
	resp, err := client.FindTeachersByFIO(ctx, &pbt.FindTeachersRequest{Fio: fio})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода FindTeachersByFIO: %v", err)
	}

	return resp.Teachers, nil
}

// Функция для поиска преподавателей по кафедре
func FindTeachersByDepartment(ctx context.Context, department string) ([]*pbt.Teacher, error) {
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbt.NewTeacherServiceClient(conn)
	resp, err := client.FindTeachersByDepartment(ctx, &pbt.FindTeachersByDepartmentRequest{Department: department})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода FindTeachersByDepartment: %v", err)
	}

	return resp.Teachers, nil
}

// Функция для поиска преподавателей по предмету
func FindTeachersBySubject(ctx context.Context, subject string) ([]*pbt.Teacher, error) {
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbt.NewTeacherServiceClient(conn)
	resp, err := client.FindTeachersBySubject(ctx, &pbt.FindTeachersBySubjectRequest{Subject: subject})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода FindTeachersBySubject: %v", err)
	}

	return resp.Teachers, nil
}

func FindAddressByPlaceName(ctx context.Context, placeName string) ([]*pbac.Place, error) {
	conn, err := grpc.Dial("localhost:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pbac.NewAddressServiceClient(conn)
	resp, err := client.GetAddressInfo(ctx, &pbac.AddressRequest{PlaceName: placeName})
	if err != nil {
		return nil, fmt.Errorf("ошибка при вызове gRPC метода GetAddressInfo: %v", err)
	}

	return resp.Places, nil
}
