// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: o5/registry/builder/v1/topic/builder.proto

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

// BuilderRequestTopicClient is the client API for BuilderRequestTopic service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BuilderRequestTopicClient interface {
	BuildProto(ctx context.Context, in *BuildProtoMessage, opts ...grpc.CallOption) (*emptypb.Empty, error)
	BuildAPI(ctx context.Context, in *BuildAPIMessage, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type builderRequestTopicClient struct {
	cc grpc.ClientConnInterface
}

func NewBuilderRequestTopicClient(cc grpc.ClientConnInterface) BuilderRequestTopicClient {
	return &builderRequestTopicClient{cc}
}

func (c *builderRequestTopicClient) BuildProto(ctx context.Context, in *BuildProtoMessage, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/o5.registry.builder.v1.topic.BuilderRequestTopic/BuildProto", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *builderRequestTopicClient) BuildAPI(ctx context.Context, in *BuildAPIMessage, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/o5.registry.builder.v1.topic.BuilderRequestTopic/BuildAPI", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BuilderRequestTopicServer is the server API for BuilderRequestTopic service.
// All implementations must embed UnimplementedBuilderRequestTopicServer
// for forward compatibility
type BuilderRequestTopicServer interface {
	BuildProto(context.Context, *BuildProtoMessage) (*emptypb.Empty, error)
	BuildAPI(context.Context, *BuildAPIMessage) (*emptypb.Empty, error)
	mustEmbedUnimplementedBuilderRequestTopicServer()
}

// UnimplementedBuilderRequestTopicServer must be embedded to have forward compatible implementations.
type UnimplementedBuilderRequestTopicServer struct {
}

func (UnimplementedBuilderRequestTopicServer) BuildProto(context.Context, *BuildProtoMessage) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuildProto not implemented")
}
func (UnimplementedBuilderRequestTopicServer) BuildAPI(context.Context, *BuildAPIMessage) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuildAPI not implemented")
}
func (UnimplementedBuilderRequestTopicServer) mustEmbedUnimplementedBuilderRequestTopicServer() {}

// UnsafeBuilderRequestTopicServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BuilderRequestTopicServer will
// result in compilation errors.
type UnsafeBuilderRequestTopicServer interface {
	mustEmbedUnimplementedBuilderRequestTopicServer()
}

func RegisterBuilderRequestTopicServer(s grpc.ServiceRegistrar, srv BuilderRequestTopicServer) {
	s.RegisterService(&BuilderRequestTopic_ServiceDesc, srv)
}

func _BuilderRequestTopic_BuildProto_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuildProtoMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BuilderRequestTopicServer).BuildProto(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/o5.registry.builder.v1.topic.BuilderRequestTopic/BuildProto",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BuilderRequestTopicServer).BuildProto(ctx, req.(*BuildProtoMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _BuilderRequestTopic_BuildAPI_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuildAPIMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BuilderRequestTopicServer).BuildAPI(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/o5.registry.builder.v1.topic.BuilderRequestTopic/BuildAPI",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BuilderRequestTopicServer).BuildAPI(ctx, req.(*BuildAPIMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// BuilderRequestTopic_ServiceDesc is the grpc.ServiceDesc for BuilderRequestTopic service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BuilderRequestTopic_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "o5.registry.builder.v1.topic.BuilderRequestTopic",
	HandlerType: (*BuilderRequestTopicServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BuildProto",
			Handler:    _BuilderRequestTopic_BuildProto_Handler,
		},
		{
			MethodName: "BuildAPI",
			Handler:    _BuilderRequestTopic_BuildAPI_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "o5/registry/builder/v1/topic/builder.proto",
}

// BuilderReplyTopicClient is the client API for BuilderReplyTopic service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BuilderReplyTopicClient interface {
	BuildStatus(ctx context.Context, in *BuildStatusMessage, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type builderReplyTopicClient struct {
	cc grpc.ClientConnInterface
}

func NewBuilderReplyTopicClient(cc grpc.ClientConnInterface) BuilderReplyTopicClient {
	return &builderReplyTopicClient{cc}
}

func (c *builderReplyTopicClient) BuildStatus(ctx context.Context, in *BuildStatusMessage, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/o5.registry.builder.v1.topic.BuilderReplyTopic/BuildStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BuilderReplyTopicServer is the server API for BuilderReplyTopic service.
// All implementations must embed UnimplementedBuilderReplyTopicServer
// for forward compatibility
type BuilderReplyTopicServer interface {
	BuildStatus(context.Context, *BuildStatusMessage) (*emptypb.Empty, error)
	mustEmbedUnimplementedBuilderReplyTopicServer()
}

// UnimplementedBuilderReplyTopicServer must be embedded to have forward compatible implementations.
type UnimplementedBuilderReplyTopicServer struct {
}

func (UnimplementedBuilderReplyTopicServer) BuildStatus(context.Context, *BuildStatusMessage) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuildStatus not implemented")
}
func (UnimplementedBuilderReplyTopicServer) mustEmbedUnimplementedBuilderReplyTopicServer() {}

// UnsafeBuilderReplyTopicServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BuilderReplyTopicServer will
// result in compilation errors.
type UnsafeBuilderReplyTopicServer interface {
	mustEmbedUnimplementedBuilderReplyTopicServer()
}

func RegisterBuilderReplyTopicServer(s grpc.ServiceRegistrar, srv BuilderReplyTopicServer) {
	s.RegisterService(&BuilderReplyTopic_ServiceDesc, srv)
}

func _BuilderReplyTopic_BuildStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuildStatusMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BuilderReplyTopicServer).BuildStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/o5.registry.builder.v1.topic.BuilderReplyTopic/BuildStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BuilderReplyTopicServer).BuildStatus(ctx, req.(*BuildStatusMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// BuilderReplyTopic_ServiceDesc is the grpc.ServiceDesc for BuilderReplyTopic service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BuilderReplyTopic_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "o5.registry.builder.v1.topic.BuilderReplyTopic",
	HandlerType: (*BuilderReplyTopicServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BuildStatus",
			Handler:    _BuilderReplyTopic_BuildStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "o5/registry/builder/v1/topic/builder.proto",
}