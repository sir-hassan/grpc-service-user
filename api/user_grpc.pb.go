// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: api/user.proto

package api

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

// UserStoreClient is the client API for UserStore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserStoreClient interface {
	CheckHealth(ctx context.Context, in *CheckHealthRequest, opts ...grpc.CallOption) (*CheckHealthReply, error)
}

type userStoreClient struct {
	cc grpc.ClientConnInterface
}

func NewUserStoreClient(cc grpc.ClientConnInterface) UserStoreClient {
	return &userStoreClient{cc}
}

func (c *userStoreClient) CheckHealth(ctx context.Context, in *CheckHealthRequest, opts ...grpc.CallOption) (*CheckHealthReply, error) {
	out := new(CheckHealthReply)
	err := c.cc.Invoke(ctx, "/api.UserStore/CheckHealth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserStoreServer is the server API for UserStore service.
// All implementations must embed UnimplementedUserStoreServer
// for forward compatibility
type UserStoreServer interface {
	CheckHealth(context.Context, *CheckHealthRequest) (*CheckHealthReply, error)
	mustEmbedUnimplementedUserStoreServer()
}

// UnimplementedUserStoreServer must be embedded to have forward compatible implementations.
type UnimplementedUserStoreServer struct {
}

func (UnimplementedUserStoreServer) CheckHealth(context.Context, *CheckHealthRequest) (*CheckHealthReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckHealth not implemented")
}
func (UnimplementedUserStoreServer) mustEmbedUnimplementedUserStoreServer() {}

// UnsafeUserStoreServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserStoreServer will
// result in compilation errors.
type UnsafeUserStoreServer interface {
	mustEmbedUnimplementedUserStoreServer()
}

func RegisterUserStoreServer(s grpc.ServiceRegistrar, srv UserStoreServer) {
	s.RegisterService(&UserStore_ServiceDesc, srv)
}

func _UserStore_CheckHealth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckHealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserStoreServer).CheckHealth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.UserStore/CheckHealth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserStoreServer).CheckHealth(ctx, req.(*CheckHealthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UserStore_ServiceDesc is the grpc.ServiceDesc for UserStore service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserStore_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.UserStore",
	HandlerType: (*UserStoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckHealth",
			Handler:    _UserStore_CheckHealth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/user.proto",
}
