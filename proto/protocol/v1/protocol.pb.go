// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.21.12
// source: protocol/v1/protocol.proto

package protocolV1

import (
	v11 "go.temporal.io/api/common/v1"
	v1 "go.temporal.io/api/failure/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Frame struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Messages []*Message `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *Frame) Reset() {
	*x = Frame{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_v1_protocol_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Frame) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Frame) ProtoMessage() {}

func (x *Frame) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_v1_protocol_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Frame.ProtoReflect.Descriptor instead.
func (*Frame) Descriptor() ([]byte, []int) {
	return file_protocol_v1_protocol_proto_rawDescGZIP(), []int{0}
}

func (x *Frame) GetMessages() []*Message {
	if x != nil {
		return x.Messages
	}
	return nil
}

// Single communication message.
type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// unique ID of the message (counter)
	Id uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// command name (if any)
	Command string `protobuf:"bytes,2,opt,name=command,proto3" json:"command,omitempty"`
	// command options in json format.
	Options []byte `protobuf:"bytes,3,opt,name=options,proto3" json:"options,omitempty"`
	// error response.
	Failure *v1.Failure `protobuf:"bytes,4,opt,name=failure,proto3" json:"failure,omitempty"`
	// invocation or result payloads.
	Payloads *v11.Payloads `protobuf:"bytes,5,opt,name=payloads,proto3" json:"payloads,omitempty"`
	// invocation or result payloads.
	Header *v11.Header `protobuf:"bytes,6,opt,name=header,proto3" json:"header,omitempty"`
	// workflow history length
	HistoryLength int64 `protobuf:"varint,7,opt,name=history_length,json=historyLength,proto3" json:"history_length,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_v1_protocol_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_v1_protocol_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_protocol_v1_protocol_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Message) GetCommand() string {
	if x != nil {
		return x.Command
	}
	return ""
}

func (x *Message) GetOptions() []byte {
	if x != nil {
		return x.Options
	}
	return nil
}

func (x *Message) GetFailure() *v1.Failure {
	if x != nil {
		return x.Failure
	}
	return nil
}

func (x *Message) GetPayloads() *v11.Payloads {
	if x != nil {
		return x.Payloads
	}
	return nil
}

func (x *Message) GetHeader() *v11.Header {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *Message) GetHistoryLength() int64 {
	if x != nil {
		return x.HistoryLength
	}
	return 0
}

var File_protocol_v1_protocol_proto protoreflect.FileDescriptor

var file_protocol_v1_protocol_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1c, 0x74, 0x65,
	0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x72, 0x6f, 0x61, 0x64, 0x72, 0x75, 0x6e, 0x6e, 0x65,
	0x72, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x1a, 0x24, 0x74, 0x65, 0x6d, 0x70,
	0x6f, 0x72, 0x61, 0x6c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f,
	0x76, 0x31, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x25, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66,
	0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4a, 0x0a, 0x05, 0x46, 0x72, 0x61, 0x6d, 0x65,
	0x12, 0x41, 0x0a, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x25, 0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x72, 0x6f,
	0x61, 0x64, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x73, 0x22, 0xa6, 0x02, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x12, 0x3a, 0x0a, 0x07, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x46,
	0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x52, 0x07, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x12,
	0x3c, 0x0a, 0x08, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x20, 0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x73, 0x52, 0x08, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x73, 0x12, 0x36, 0x0a,
	0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e,
	0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x25, 0x0a, 0x0e, 0x68, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79,
	0x5f, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x68,
	0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x42, 0x0f, 0x5a, 0x0d,
	0x2e, 0x2f, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protocol_v1_protocol_proto_rawDescOnce sync.Once
	file_protocol_v1_protocol_proto_rawDescData = file_protocol_v1_protocol_proto_rawDesc
)

func file_protocol_v1_protocol_proto_rawDescGZIP() []byte {
	file_protocol_v1_protocol_proto_rawDescOnce.Do(func() {
		file_protocol_v1_protocol_proto_rawDescData = protoimpl.X.CompressGZIP(file_protocol_v1_protocol_proto_rawDescData)
	})
	return file_protocol_v1_protocol_proto_rawDescData
}

var file_protocol_v1_protocol_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protocol_v1_protocol_proto_goTypes = []interface{}{
	(*Frame)(nil),        // 0: temporal.roadrunner.internal.Frame
	(*Message)(nil),      // 1: temporal.roadrunner.internal.Message
	(*v1.Failure)(nil),   // 2: temporal.api.failure.v1.Failure
	(*v11.Payloads)(nil), // 3: temporal.api.common.v1.Payloads
	(*v11.Header)(nil),   // 4: temporal.api.common.v1.Header
}
var file_protocol_v1_protocol_proto_depIdxs = []int32{
	1, // 0: temporal.roadrunner.internal.Frame.messages:type_name -> temporal.roadrunner.internal.Message
	2, // 1: temporal.roadrunner.internal.Message.failure:type_name -> temporal.api.failure.v1.Failure
	3, // 2: temporal.roadrunner.internal.Message.payloads:type_name -> temporal.api.common.v1.Payloads
	4, // 3: temporal.roadrunner.internal.Message.header:type_name -> temporal.api.common.v1.Header
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_protocol_v1_protocol_proto_init() }
func file_protocol_v1_protocol_proto_init() {
	if File_protocol_v1_protocol_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protocol_v1_protocol_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Frame); i {
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
		file_protocol_v1_protocol_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
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
			RawDescriptor: file_protocol_v1_protocol_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protocol_v1_protocol_proto_goTypes,
		DependencyIndexes: file_protocol_v1_protocol_proto_depIdxs,
		MessageInfos:      file_protocol_v1_protocol_proto_msgTypes,
	}.Build()
	File_protocol_v1_protocol_proto = out.File
	file_protocol_v1_protocol_proto_rawDesc = nil
	file_protocol_v1_protocol_proto_goTypes = nil
	file_protocol_v1_protocol_proto_depIdxs = nil
}
