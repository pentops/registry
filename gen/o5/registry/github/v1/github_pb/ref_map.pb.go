// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: o5/registry/github/v1/ref_map.proto

package github_pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DeployerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Refs []*RefLink `protobuf:"bytes,1,rep,name=refs,proto3" json:"refs,omitempty"`
	// source the named env files, s3:// or local file format
	TargetEnvironments []string `protobuf:"bytes,2,rep,name=target_environments,json=targetEnvironments,proto3" json:"target_environments,omitempty"`
}

func (x *DeployerConfig) Reset() {
	*x = DeployerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_o5_registry_github_v1_ref_map_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeployerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeployerConfig) ProtoMessage() {}

func (x *DeployerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_o5_registry_github_v1_ref_map_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeployerConfig.ProtoReflect.Descriptor instead.
func (*DeployerConfig) Descriptor() ([]byte, []int) {
	return file_o5_registry_github_v1_ref_map_proto_rawDescGZIP(), []int{0}
}

func (x *DeployerConfig) GetRefs() []*RefLink {
	if x != nil {
		return x.Refs
	}
	return nil
}

func (x *DeployerConfig) GetTargetEnvironments() []string {
	if x != nil {
		return x.TargetEnvironments
	}
	return nil
}

type RefLink struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Owner    string   `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Repo     string   `protobuf:"bytes,2,opt,name=repo,proto3" json:"repo,omitempty"`
	RefMatch string   `protobuf:"bytes,3,opt,name=ref_match,json=refMatch,proto3" json:"ref_match,omitempty"` // refs/heads/main or refs/tags/* etc, using wildcards.
	Targets  []string `protobuf:"bytes,4,rep,name=targets,proto3" json:"targets,omitempty"`                   // the name matching a config file
}

func (x *RefLink) Reset() {
	*x = RefLink{}
	if protoimpl.UnsafeEnabled {
		mi := &file_o5_registry_github_v1_ref_map_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RefLink) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RefLink) ProtoMessage() {}

func (x *RefLink) ProtoReflect() protoreflect.Message {
	mi := &file_o5_registry_github_v1_ref_map_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RefLink.ProtoReflect.Descriptor instead.
func (*RefLink) Descriptor() ([]byte, []int) {
	return file_o5_registry_github_v1_ref_map_proto_rawDescGZIP(), []int{1}
}

func (x *RefLink) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *RefLink) GetRepo() string {
	if x != nil {
		return x.Repo
	}
	return ""
}

func (x *RefLink) GetRefMatch() string {
	if x != nil {
		return x.RefMatch
	}
	return ""
}

func (x *RefLink) GetTargets() []string {
	if x != nil {
		return x.Targets
	}
	return nil
}

var File_o5_registry_github_v1_ref_map_proto protoreflect.FileDescriptor

var file_o5_registry_github_v1_ref_map_proto_rawDesc = []byte{
	0x0a, 0x23, 0x6f, 0x35, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x66, 0x5f, 0x6d, 0x61, 0x70, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x6f, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x72, 0x79, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x1a, 0x20, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x75,
	0x0a, 0x0e, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x32, 0x0a, 0x04, 0x72, 0x65, 0x66, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e,
	0x2e, 0x6f, 0x35, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x66, 0x4c, 0x69, 0x6e, 0x6b, 0x52, 0x04,
	0x72, 0x65, 0x66, 0x73, 0x12, 0x2f, 0x0a, 0x13, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x65,
	0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x12, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x45, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e,
	0x6d, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x6a, 0x0a, 0x07, 0x52, 0x65, 0x66, 0x4c, 0x69, 0x6e, 0x6b,
	0x12, 0x14, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x12, 0x1b, 0x0a, 0x09, 0x72, 0x65,
	0x66, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72,
	0x65, 0x66, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74,
	0x73, 0x42, 0x41, 0x5a, 0x3f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6f, 0x35, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x5f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_o5_registry_github_v1_ref_map_proto_rawDescOnce sync.Once
	file_o5_registry_github_v1_ref_map_proto_rawDescData = file_o5_registry_github_v1_ref_map_proto_rawDesc
)

func file_o5_registry_github_v1_ref_map_proto_rawDescGZIP() []byte {
	file_o5_registry_github_v1_ref_map_proto_rawDescOnce.Do(func() {
		file_o5_registry_github_v1_ref_map_proto_rawDescData = protoimpl.X.CompressGZIP(file_o5_registry_github_v1_ref_map_proto_rawDescData)
	})
	return file_o5_registry_github_v1_ref_map_proto_rawDescData
}

var file_o5_registry_github_v1_ref_map_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_o5_registry_github_v1_ref_map_proto_goTypes = []interface{}{
	(*DeployerConfig)(nil), // 0: o5.registry.github.v1.DeployerConfig
	(*RefLink)(nil),        // 1: o5.registry.github.v1.RefLink
}
var file_o5_registry_github_v1_ref_map_proto_depIdxs = []int32{
	1, // 0: o5.registry.github.v1.DeployerConfig.refs:type_name -> o5.registry.github.v1.RefLink
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_o5_registry_github_v1_ref_map_proto_init() }
func file_o5_registry_github_v1_ref_map_proto_init() {
	if File_o5_registry_github_v1_ref_map_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_o5_registry_github_v1_ref_map_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeployerConfig); i {
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
		file_o5_registry_github_v1_ref_map_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RefLink); i {
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
			RawDescriptor: file_o5_registry_github_v1_ref_map_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_o5_registry_github_v1_ref_map_proto_goTypes,
		DependencyIndexes: file_o5_registry_github_v1_ref_map_proto_depIdxs,
		MessageInfos:      file_o5_registry_github_v1_ref_map_proto_msgTypes,
	}.Build()
	File_o5_registry_github_v1_ref_map_proto = out.File
	file_o5_registry_github_v1_ref_map_proto_rawDesc = nil
	file_o5_registry_github_v1_ref_map_proto_goTypes = nil
	file_o5_registry_github_v1_ref_map_proto_depIdxs = nil
}
