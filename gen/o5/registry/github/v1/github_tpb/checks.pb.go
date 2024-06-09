// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: o5/registry/github/v1/topic/checks.proto

package github_tpb

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	_ "github.com/pentops/o5-go/messaging/v1/messaging_pb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
	_ "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CheckRun struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Owner     string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Repo      string `protobuf:"bytes,2,opt,name=repo,proto3" json:"repo,omitempty"`
	CheckName string `protobuf:"bytes,3,opt,name=check_name,json=checkName,proto3" json:"check_name,omitempty"`
	CheckId   int64  `protobuf:"varint,4,opt,name=check_id,json=checkId,proto3" json:"check_id,omitempty"`
}

func (x *CheckRun) Reset() {
	*x = CheckRun{}
	if protoimpl.UnsafeEnabled {
		mi := &file_o5_registry_github_v1_topic_checks_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckRun) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckRun) ProtoMessage() {}

func (x *CheckRun) ProtoReflect() protoreflect.Message {
	mi := &file_o5_registry_github_v1_topic_checks_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckRun.ProtoReflect.Descriptor instead.
func (*CheckRun) Descriptor() ([]byte, []int) {
	return file_o5_registry_github_v1_topic_checks_proto_rawDescGZIP(), []int{0}
}

func (x *CheckRun) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *CheckRun) GetRepo() string {
	if x != nil {
		return x.Repo
	}
	return ""
}

func (x *CheckRun) GetCheckName() string {
	if x != nil {
		return x.CheckName
	}
	return ""
}

func (x *CheckRun) GetCheckId() int64 {
	if x != nil {
		return x.CheckId
	}
	return 0
}

var File_o5_registry_github_v1_topic_checks_proto protoreflect.FileDescriptor

var file_o5_registry_github_v1_topic_checks_proto_rawDesc = []byte{
	0x0a, 0x28, 0x6f, 0x35, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x2f, 0x63, 0x68,
	0x65, 0x63, 0x6b, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1b, 0x6f, 0x35, 0x2e, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x76,
	0x31, 0x2e, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x6f, 0x35, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x69, 0x6e,
	0x67, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6f, 0x35, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x71, 0x72, 0x65, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x6e, 0x0a, 0x08, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x75, 0x6e,
	0x12, 0x14, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x68,
	0x65, 0x63, 0x6b, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x63, 0x68, 0x65, 0x63, 0x6b, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x68, 0x65,
	0x63, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x63, 0x68, 0x65,
	0x63, 0x6b, 0x49, 0x64, 0x42, 0x42, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x70, 0x65, 0x6e, 0x74, 0x6f, 0x70, 0x73, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6f, 0x35, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x5f, 0x74, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_o5_registry_github_v1_topic_checks_proto_rawDescOnce sync.Once
	file_o5_registry_github_v1_topic_checks_proto_rawDescData = file_o5_registry_github_v1_topic_checks_proto_rawDesc
)

func file_o5_registry_github_v1_topic_checks_proto_rawDescGZIP() []byte {
	file_o5_registry_github_v1_topic_checks_proto_rawDescOnce.Do(func() {
		file_o5_registry_github_v1_topic_checks_proto_rawDescData = protoimpl.X.CompressGZIP(file_o5_registry_github_v1_topic_checks_proto_rawDescData)
	})
	return file_o5_registry_github_v1_topic_checks_proto_rawDescData
}

var file_o5_registry_github_v1_topic_checks_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_o5_registry_github_v1_topic_checks_proto_goTypes = []interface{}{
	(*CheckRun)(nil), // 0: o5.registry.github.v1.topic.CheckRun
}
var file_o5_registry_github_v1_topic_checks_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_o5_registry_github_v1_topic_checks_proto_init() }
func file_o5_registry_github_v1_topic_checks_proto_init() {
	if File_o5_registry_github_v1_topic_checks_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_o5_registry_github_v1_topic_checks_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckRun); i {
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
			RawDescriptor: file_o5_registry_github_v1_topic_checks_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_o5_registry_github_v1_topic_checks_proto_goTypes,
		DependencyIndexes: file_o5_registry_github_v1_topic_checks_proto_depIdxs,
		MessageInfos:      file_o5_registry_github_v1_topic_checks_proto_msgTypes,
	}.Build()
	File_o5_registry_github_v1_topic_checks_proto = out.File
	file_o5_registry_github_v1_topic_checks_proto_rawDesc = nil
	file_o5_registry_github_v1_topic_checks_proto_goTypes = nil
	file_o5_registry_github_v1_topic_checks_proto_depIdxs = nil
}
