// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: j5/registry/github/v1/service/repo_command.proto

package github_spb

import (
	reflect "reflect"
	sync "sync"

	_ "github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	github_pb "github.com/pentops/registry/internal/gen/j5/registry/github/v1/github_pb"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ConfigureRepoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Owner  string                             `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Name   string                             `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Config *github_pb.RepoEventType_Configure `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *ConfigureRepoRequest) Reset() {
	*x = ConfigureRepoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_registry_github_v1_service_repo_command_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfigureRepoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigureRepoRequest) ProtoMessage() {}

func (x *ConfigureRepoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_j5_registry_github_v1_service_repo_command_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigureRepoRequest.ProtoReflect.Descriptor instead.
func (*ConfigureRepoRequest) Descriptor() ([]byte, []int) {
	return file_j5_registry_github_v1_service_repo_command_proto_rawDescGZIP(), []int{0}
}

func (x *ConfigureRepoRequest) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *ConfigureRepoRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ConfigureRepoRequest) GetConfig() *github_pb.RepoEventType_Configure {
	if x != nil {
		return x.Config
	}
	return nil
}

type ConfigureRepoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Repo *github_pb.RepoState `protobuf:"bytes,1,opt,name=repo,proto3" json:"repo,omitempty"`
}

func (x *ConfigureRepoResponse) Reset() {
	*x = ConfigureRepoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_registry_github_v1_service_repo_command_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfigureRepoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfigureRepoResponse) ProtoMessage() {}

func (x *ConfigureRepoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_j5_registry_github_v1_service_repo_command_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfigureRepoResponse.ProtoReflect.Descriptor instead.
func (*ConfigureRepoResponse) Descriptor() ([]byte, []int) {
	return file_j5_registry_github_v1_service_repo_command_proto_rawDescGZIP(), []int{1}
}

func (x *ConfigureRepoResponse) GetRepo() *github_pb.RepoState {
	if x != nil {
		return x.Repo
	}
	return nil
}

type TriggerRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Owner  string                      `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Repo   string                      `protobuf:"bytes,2,opt,name=repo,proto3" json:"repo,omitempty"`
	Commit string                      `protobuf:"bytes,3,opt,name=commit,proto3" json:"commit,omitempty"`
	Target *github_pb.DeployTargetType `protobuf:"bytes,4,opt,name=target,proto3" json:"target,omitempty"`
}

func (x *TriggerRequest) Reset() {
	*x = TriggerRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_registry_github_v1_service_repo_command_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TriggerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TriggerRequest) ProtoMessage() {}

func (x *TriggerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_j5_registry_github_v1_service_repo_command_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TriggerRequest.ProtoReflect.Descriptor instead.
func (*TriggerRequest) Descriptor() ([]byte, []int) {
	return file_j5_registry_github_v1_service_repo_command_proto_rawDescGZIP(), []int{2}
}

func (x *TriggerRequest) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *TriggerRequest) GetRepo() string {
	if x != nil {
		return x.Repo
	}
	return ""
}

func (x *TriggerRequest) GetCommit() string {
	if x != nil {
		return x.Commit
	}
	return ""
}

func (x *TriggerRequest) GetTarget() *github_pb.DeployTargetType {
	if x != nil {
		return x.Target
	}
	return nil
}

type TriggerResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TriggerResponse) Reset() {
	*x = TriggerResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_j5_registry_github_v1_service_repo_command_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TriggerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TriggerResponse) ProtoMessage() {}

func (x *TriggerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_j5_registry_github_v1_service_repo_command_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TriggerResponse.ProtoReflect.Descriptor instead.
func (*TriggerResponse) Descriptor() ([]byte, []int) {
	return file_j5_registry_github_v1_service_repo_command_proto_rawDescGZIP(), []int{3}
}

var File_j5_registry_github_v1_service_repo_command_proto protoreflect.FileDescriptor

var file_j5_registry_github_v1_service_repo_command_proto_rawDesc = []byte{
	0x0a, 0x30, 0x6a, 0x35, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f,
	0x72, 0x65, 0x70, 0x6f, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x1d, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1b, 0x6a, 0x35, 0x2f, 0x65, 0x78, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x6a, 0x35,
	0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x70, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x88,
	0x01, 0x0a, 0x14, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x52, 0x65, 0x70, 0x6f,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x46, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x2e, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72,
	0x65, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x4d, 0x0a, 0x15, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x34, 0x0a, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x20, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x52, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x22, 0x93, 0x01, 0x0a, 0x0e, 0x54, 0x72, 0x69,
	0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6f,
	0x77, 0x6e, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65,
	0x72, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x72, 0x65, 0x70, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x3f, 0x0a,
	0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e,
	0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x54, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x06, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x22, 0x11,
	0x0a, 0x0f, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x32, 0x89, 0x03, 0x0a, 0x12, 0x52, 0x65, 0x70, 0x6f, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0xba, 0x01, 0x0a, 0x0d, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x12, 0x33, 0x2e, 0x6a, 0x35, 0x2e,
	0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x76, 0x31, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x75, 0x72, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x34, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75, 0x72, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x3e, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x38, 0x3a, 0x01, 0x2a,
	0x22, 0x33, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x2f, 0x72, 0x65, 0x70, 0x6f, 0x2f, 0x7b, 0x6f, 0x77,
	0x6e, 0x65, 0x72, 0x7d, 0x2f, 0x7b, 0x6e, 0x61, 0x6d, 0x65, 0x7d, 0x2f, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x75, 0x72, 0x65, 0x12, 0xa6, 0x01, 0x0a, 0x07, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65,
	0x72, 0x12, 0x2d, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x2e, 0x2e, 0x6a, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x3c, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x36, 0x3a, 0x01, 0x2a, 0x22, 0x31, 0x2f, 0x72, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31,
	0x2f, 0x63, 0x2f, 0x72, 0x65, 0x70, 0x6f, 0x2f, 0x7b, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x7d, 0x2f,
	0x7b, 0x72, 0x65, 0x70, 0x6f, 0x7d, 0x2f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x1a, 0x0d,
	0xea, 0x85, 0x8f, 0x02, 0x08, 0x12, 0x06, 0x0a, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x42, 0x4b, 0x5a,
	0x49, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74,
	0x6f, 0x70, 0x73, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6a, 0x35, 0x2f, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x5f, 0x73, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_j5_registry_github_v1_service_repo_command_proto_rawDescOnce sync.Once
	file_j5_registry_github_v1_service_repo_command_proto_rawDescData = file_j5_registry_github_v1_service_repo_command_proto_rawDesc
)

func file_j5_registry_github_v1_service_repo_command_proto_rawDescGZIP() []byte {
	file_j5_registry_github_v1_service_repo_command_proto_rawDescOnce.Do(func() {
		file_j5_registry_github_v1_service_repo_command_proto_rawDescData = protoimpl.X.CompressGZIP(file_j5_registry_github_v1_service_repo_command_proto_rawDescData)
	})
	return file_j5_registry_github_v1_service_repo_command_proto_rawDescData
}

var file_j5_registry_github_v1_service_repo_command_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_j5_registry_github_v1_service_repo_command_proto_goTypes = []interface{}{
	(*ConfigureRepoRequest)(nil),              // 0: j5.registry.github.v1.service.ConfigureRepoRequest
	(*ConfigureRepoResponse)(nil),             // 1: j5.registry.github.v1.service.ConfigureRepoResponse
	(*TriggerRequest)(nil),                    // 2: j5.registry.github.v1.service.TriggerRequest
	(*TriggerResponse)(nil),                   // 3: j5.registry.github.v1.service.TriggerResponse
	(*github_pb.RepoEventType_Configure)(nil), // 4: j5.registry.github.v1.RepoEventType.Configure
	(*github_pb.RepoState)(nil),               // 5: j5.registry.github.v1.RepoState
	(*github_pb.DeployTargetType)(nil),        // 6: j5.registry.github.v1.DeployTargetType
}
var file_j5_registry_github_v1_service_repo_command_proto_depIdxs = []int32{
	4, // 0: j5.registry.github.v1.service.ConfigureRepoRequest.config:type_name -> j5.registry.github.v1.RepoEventType.Configure
	5, // 1: j5.registry.github.v1.service.ConfigureRepoResponse.repo:type_name -> j5.registry.github.v1.RepoState
	6, // 2: j5.registry.github.v1.service.TriggerRequest.target:type_name -> j5.registry.github.v1.DeployTargetType
	0, // 3: j5.registry.github.v1.service.RepoCommandService.ConfigureRepo:input_type -> j5.registry.github.v1.service.ConfigureRepoRequest
	2, // 4: j5.registry.github.v1.service.RepoCommandService.Trigger:input_type -> j5.registry.github.v1.service.TriggerRequest
	1, // 5: j5.registry.github.v1.service.RepoCommandService.ConfigureRepo:output_type -> j5.registry.github.v1.service.ConfigureRepoResponse
	3, // 6: j5.registry.github.v1.service.RepoCommandService.Trigger:output_type -> j5.registry.github.v1.service.TriggerResponse
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_j5_registry_github_v1_service_repo_command_proto_init() }
func file_j5_registry_github_v1_service_repo_command_proto_init() {
	if File_j5_registry_github_v1_service_repo_command_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_j5_registry_github_v1_service_repo_command_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConfigureRepoRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_j5_registry_github_v1_service_repo_command_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConfigureRepoResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_j5_registry_github_v1_service_repo_command_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TriggerRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_j5_registry_github_v1_service_repo_command_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TriggerResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_j5_registry_github_v1_service_repo_command_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_j5_registry_github_v1_service_repo_command_proto_goTypes,
		DependencyIndexes: file_j5_registry_github_v1_service_repo_command_proto_depIdxs,
		MessageInfos:      file_j5_registry_github_v1_service_repo_command_proto_msgTypes,
	}.Build()
	File_j5_registry_github_v1_service_repo_command_proto = out.File
	file_j5_registry_github_v1_service_repo_command_proto_rawDesc = nil
	file_j5_registry_github_v1_service_repo_command_proto_goTypes = nil
	file_j5_registry_github_v1_service_repo_command_proto_depIdxs = nil
}
