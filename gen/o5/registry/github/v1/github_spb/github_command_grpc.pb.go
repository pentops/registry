// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: o5/registry/github/v1/service/github_command.proto

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

// GithubCommandServiceClient is the client API for GithubCommandService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GithubCommandServiceClient interface {
	ConfigureRepo(ctx context.Context, in *ConfigureRepoRequest, opts ...grpc.CallOption) (*ConfigureRepoResponse, error)
}

type githubCommandServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGithubCommandServiceClient(cc grpc.ClientConnInterface) GithubCommandServiceClient {
	return &githubCommandServiceClient{cc}
}

func (c *githubCommandServiceClient) ConfigureRepo(ctx context.Context, in *ConfigureRepoRequest, opts ...grpc.CallOption) (*ConfigureRepoResponse, error) {
	out := new(ConfigureRepoResponse)
	err := c.cc.Invoke(ctx, "/o5.registry.github.v1.service.GithubCommandService/ConfigureRepo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GithubCommandServiceServer is the server API for GithubCommandService service.
// All implementations must embed UnimplementedGithubCommandServiceServer
// for forward compatibility
type GithubCommandServiceServer interface {
	ConfigureRepo(context.Context, *ConfigureRepoRequest) (*ConfigureRepoResponse, error)
	mustEmbedUnimplementedGithubCommandServiceServer()
}

// UnimplementedGithubCommandServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGithubCommandServiceServer struct {
}

func (UnimplementedGithubCommandServiceServer) ConfigureRepo(context.Context, *ConfigureRepoRequest) (*ConfigureRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfigureRepo not implemented")
}
func (UnimplementedGithubCommandServiceServer) mustEmbedUnimplementedGithubCommandServiceServer() {}

// UnsafeGithubCommandServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GithubCommandServiceServer will
// result in compilation errors.
type UnsafeGithubCommandServiceServer interface {
	mustEmbedUnimplementedGithubCommandServiceServer()
}

func RegisterGithubCommandServiceServer(s grpc.ServiceRegistrar, srv GithubCommandServiceServer) {
	s.RegisterService(&GithubCommandService_ServiceDesc, srv)
}

func _GithubCommandService_ConfigureRepo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfigureRepoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GithubCommandServiceServer).ConfigureRepo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/o5.registry.github.v1.service.GithubCommandService/ConfigureRepo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GithubCommandServiceServer).ConfigureRepo(ctx, req.(*ConfigureRepoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GithubCommandService_ServiceDesc is the grpc.ServiceDesc for GithubCommandService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GithubCommandService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "o5.registry.github.v1.service.GithubCommandService",
	HandlerType: (*GithubCommandServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ConfigureRepo",
			Handler:    _GithubCommandService_ConfigureRepo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "o5/registry/github/v1/service/github_command.proto",
}