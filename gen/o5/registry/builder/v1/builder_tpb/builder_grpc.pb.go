// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: o5/registry/builder/v1/builder.proto

package builder_tpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// BuilderTopicClient is the client API for BuilderTopic service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BuilderTopicClient interface {
	BuildProto(ctx context.Context, in *BuildProtoMessage, opts ...grpc.CallOption) (*emptypb.Empty, error)
	BuildAPI(ctx context.Context, in *BuildAPIMessage, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type builderTopicClient struct {
	cc grpc.ClientConnInterface
}

func NewBuilderTopicClient(cc grpc.ClientConnInterface) BuilderTopicClient {
	return &builderTopicClient{cc}
}

func (c *builderTopicClient) BuildProto(ctx context.Context, in *BuildProtoMessage, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/o5.registry.builder.v1.BuilderTopic/BuildProto", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *builderTopicClient) BuildAPI(ctx context.Context, in *BuildAPIMessage, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/o5.registry.builder.v1.BuilderTopic/BuildAPI", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BuilderTopicServer is the server API for BuilderTopic service.
// All implementations must embed UnimplementedBuilderTopicServer
// for forward compatibility
type BuilderTopicServer interface {
	BuildProto(context.Context, *BuildProtoMessage) (*emptypb.Empty, error)
	BuildAPI(context.Context, *BuildAPIMessage) (*emptypb.Empty, error)
	mustEmbedUnimplementedBuilderTopicServer()
}

// UnimplementedBuilderTopicServer must be embedded to have forward compatible implementations.
type UnimplementedBuilderTopicServer struct {
}

func (UnimplementedBuilderTopicServer) BuildProto(context.Context, *BuildProtoMessage) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuildProto not implemented")
}
func (UnimplementedBuilderTopicServer) BuildAPI(context.Context, *BuildAPIMessage) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuildAPI not implemented")
}
func (UnimplementedBuilderTopicServer) mustEmbedUnimplementedBuilderTopicServer() {}

// UnsafeBuilderTopicServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BuilderTopicServer will
// result in compilation errors.
type UnsafeBuilderTopicServer interface {
	mustEmbedUnimplementedBuilderTopicServer()
}

func RegisterBuilderTopicServer(s grpc.ServiceRegistrar, srv BuilderTopicServer) {
	s.RegisterService(&BuilderTopic_ServiceDesc, srv)
}

func _BuilderTopic_BuildProto_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuildProtoMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BuilderTopicServer).BuildProto(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/o5.registry.builder.v1.BuilderTopic/BuildProto",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BuilderTopicServer).BuildProto(ctx, req.(*BuildProtoMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _BuilderTopic_BuildAPI_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuildAPIMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BuilderTopicServer).BuildAPI(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/o5.registry.builder.v1.BuilderTopic/BuildAPI",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BuilderTopicServer).BuildAPI(ctx, req.(*BuildAPIMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// BuilderTopic_ServiceDesc is the grpc.ServiceDesc for BuilderTopic service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BuilderTopic_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "o5.registry.builder.v1.BuilderTopic",
	HandlerType: (*BuilderTopicServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BuildProto",
			Handler:    _BuilderTopic_BuildProto_Handler,
		},
		{
			MethodName: "BuildAPI",
			Handler:    _BuilderTopic_BuildAPI_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "o5/registry/builder/v1/builder.proto",
}
