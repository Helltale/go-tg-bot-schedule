// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v5.28.0
// source: teacher.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TeacherServiceClient is the client API for TeacherService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TeacherServiceClient interface {
	FindTeachersByFIO(ctx context.Context, in *FindTeachersRequest, opts ...grpc.CallOption) (*FindTeachersResponse, error)
	FindTeachersByDepartment(ctx context.Context, in *FindTeachersByDepartmentRequest, opts ...grpc.CallOption) (*FindTeachersResponse, error)
	FindTeachersBySubject(ctx context.Context, in *FindTeachersBySubjectRequest, opts ...grpc.CallOption) (*FindTeachersResponse, error)
}

type teacherServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTeacherServiceClient(cc grpc.ClientConnInterface) TeacherServiceClient {
	return &teacherServiceClient{cc}
}

func (c *teacherServiceClient) FindTeachersByFIO(ctx context.Context, in *FindTeachersRequest, opts ...grpc.CallOption) (*FindTeachersResponse, error) {
	out := new(FindTeachersResponse)
	err := c.cc.Invoke(ctx, "/teacher.TeacherService/FindTeachersByFIO", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *teacherServiceClient) FindTeachersByDepartment(ctx context.Context, in *FindTeachersByDepartmentRequest, opts ...grpc.CallOption) (*FindTeachersResponse, error) {
	out := new(FindTeachersResponse)
	err := c.cc.Invoke(ctx, "/teacher.TeacherService/FindTeachersByDepartment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *teacherServiceClient) FindTeachersBySubject(ctx context.Context, in *FindTeachersBySubjectRequest, opts ...grpc.CallOption) (*FindTeachersResponse, error) {
	out := new(FindTeachersResponse)
	err := c.cc.Invoke(ctx, "/teacher.TeacherService/FindTeachersBySubject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TeacherServiceServer is the server API for TeacherService service.
// All implementations must embed UnimplementedTeacherServiceServer
// for forward compatibility
type TeacherServiceServer interface {
	FindTeachersByFIO(context.Context, *FindTeachersRequest) (*FindTeachersResponse, error)
	FindTeachersByDepartment(context.Context, *FindTeachersByDepartmentRequest) (*FindTeachersResponse, error)
	FindTeachersBySubject(context.Context, *FindTeachersBySubjectRequest) (*FindTeachersResponse, error)
	mustEmbedUnimplementedTeacherServiceServer()
}

// UnimplementedTeacherServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTeacherServiceServer struct {
}

func (UnimplementedTeacherServiceServer) FindTeachersByFIO(context.Context, *FindTeachersRequest) (*FindTeachersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindTeachersByFIO not implemented")
}
func (UnimplementedTeacherServiceServer) FindTeachersByDepartment(context.Context, *FindTeachersByDepartmentRequest) (*FindTeachersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindTeachersByDepartment not implemented")
}
func (UnimplementedTeacherServiceServer) FindTeachersBySubject(context.Context, *FindTeachersBySubjectRequest) (*FindTeachersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindTeachersBySubject not implemented")
}
func (UnimplementedTeacherServiceServer) mustEmbedUnimplementedTeacherServiceServer() {}

// UnsafeTeacherServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TeacherServiceServer will
// result in compilation errors.
type UnsafeTeacherServiceServer interface {
	mustEmbedUnimplementedTeacherServiceServer()
}

func RegisterTeacherServiceServer(s grpc.ServiceRegistrar, srv TeacherServiceServer) {
	s.RegisterService(&TeacherService_ServiceDesc, srv)
}

func _TeacherService_FindTeachersByFIO_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindTeachersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TeacherServiceServer).FindTeachersByFIO(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teacher.TeacherService/FindTeachersByFIO",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TeacherServiceServer).FindTeachersByFIO(ctx, req.(*FindTeachersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TeacherService_FindTeachersByDepartment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindTeachersByDepartmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TeacherServiceServer).FindTeachersByDepartment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teacher.TeacherService/FindTeachersByDepartment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TeacherServiceServer).FindTeachersByDepartment(ctx, req.(*FindTeachersByDepartmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TeacherService_FindTeachersBySubject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindTeachersBySubjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TeacherServiceServer).FindTeachersBySubject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teacher.TeacherService/FindTeachersBySubject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TeacherServiceServer).FindTeachersBySubject(ctx, req.(*FindTeachersBySubjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TeacherService_ServiceDesc is the grpc.ServiceDesc for TeacherService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TeacherService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "teacher.TeacherService",
	HandlerType: (*TeacherServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FindTeachersByFIO",
			Handler:    _TeacherService_FindTeachersByFIO_Handler,
		},
		{
			MethodName: "FindTeachersByDepartment",
			Handler:    _TeacherService_FindTeachersByDepartment_Handler,
		},
		{
			MethodName: "FindTeachersBySubject",
			Handler:    _TeacherService_FindTeachersBySubject_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "teacher.proto",
}
