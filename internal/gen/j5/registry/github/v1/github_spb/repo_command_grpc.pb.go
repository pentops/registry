// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: j5/registry/github/v1/service/repo_command.proto

package github_spb

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

const (
	RepoCommandService_ConfigureRepo_FullMethodName = "/j5.registry.github.v1.service.RepoCommandService/ConfigureRepo"
	RepoCommandService_Trigger_FullMethodName       = "/j5.registry.github.v1.service.RepoCommandService/Trigger"
)

// RepoCommandServiceClient is the client API for RepoCommandService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RepoCommandServiceClient interface {
	ConfigureRepo(ctx context.Context, in *ConfigureRepoRequest, opts ...grpc.CallOption) (*ConfigureRepoResponse, error)
	Trigger(ctx context.Context, in *TriggerRequest, opts ...grpc.CallOption) (*TriggerResponse, error)
}

type repoCommandServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRepoCommandServiceClient(cc grpc.ClientConnInterface) RepoCommandServiceClient {
	return &repoCommandServiceClient{cc}
}

func (c *repoCommandServiceClient) ConfigureRepo(ctx context.Context, in *ConfigureRepoRequest, opts ...grpc.CallOption) (*ConfigureRepoResponse, error) {
	out := new(ConfigureRepoResponse)
	err := c.cc.Invoke(ctx, RepoCommandService_ConfigureRepo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repoCommandServiceClient) Trigger(ctx context.Context, in *TriggerRequest, opts ...grpc.CallOption) (*TriggerResponse, error) {
	out := new(TriggerResponse)
	err := c.cc.Invoke(ctx, RepoCommandService_Trigger_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RepoCommandServiceServer is the server API for RepoCommandService service.
// All implementations must embed UnimplementedRepoCommandServiceServer
// for forward compatibility
type RepoCommandServiceServer interface {
	ConfigureRepo(context.Context, *ConfigureRepoRequest) (*ConfigureRepoResponse, error)
	Trigger(context.Context, *TriggerRequest) (*TriggerResponse, error)
	mustEmbedUnimplementedRepoCommandServiceServer()
}

// UnimplementedRepoCommandServiceServer must be embedded to have forward compatible implementations.
type UnimplementedRepoCommandServiceServer struct {
}

func (UnimplementedRepoCommandServiceServer) ConfigureRepo(context.Context, *ConfigureRepoRequest) (*ConfigureRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfigureRepo not implemented")
}
func (UnimplementedRepoCommandServiceServer) Trigger(context.Context, *TriggerRequest) (*TriggerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Trigger not implemented")
}
func (UnimplementedRepoCommandServiceServer) mustEmbedUnimplementedRepoCommandServiceServer() {}

// UnsafeRepoCommandServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RepoCommandServiceServer will
// result in compilation errors.
type UnsafeRepoCommandServiceServer interface {
	mustEmbedUnimplementedRepoCommandServiceServer()
}

func RegisterRepoCommandServiceServer(s grpc.ServiceRegistrar, srv RepoCommandServiceServer) {
	s.RegisterService(&RepoCommandService_ServiceDesc, srv)
}

func _RepoCommandService_ConfigureRepo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfigureRepoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepoCommandServiceServer).ConfigureRepo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RepoCommandService_ConfigureRepo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepoCommandServiceServer).ConfigureRepo(ctx, req.(*ConfigureRepoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepoCommandService_Trigger_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TriggerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepoCommandServiceServer).Trigger(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RepoCommandService_Trigger_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepoCommandServiceServer).Trigger(ctx, req.(*TriggerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RepoCommandService_ServiceDesc is the grpc.ServiceDesc for RepoCommandService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RepoCommandService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "j5.registry.github.v1.service.RepoCommandService",
	HandlerType: (*RepoCommandServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ConfigureRepo",
			Handler:    _RepoCommandService_ConfigureRepo_Handler,
		},
		{
			MethodName: "Trigger",
			Handler:    _RepoCommandService_Trigger_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "j5/registry/github/v1/service/repo_command.proto",
}
