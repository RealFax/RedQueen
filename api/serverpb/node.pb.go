// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: node.proto

package serverpb

import (
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

type RaftLogCommand int32

const (
	RaftLogCommand_SetWithTTL    RaftLogCommand = 0
	RaftLogCommand_TrySetWithTTL RaftLogCommand = 1
	RaftLogCommand_Set           RaftLogCommand = 2
	RaftLogCommand_TrySet        RaftLogCommand = 3
	RaftLogCommand_Del           RaftLogCommand = 4
)

// Enum value maps for RaftLogCommand.
var (
	RaftLogCommand_name = map[int32]string{
		0: "SetWithTTL",
		1: "TrySetWithTTL",
		2: "Set",
		3: "TrySet",
		4: "Del",
	}
	RaftLogCommand_value = map[string]int32{
		"SetWithTTL":    0,
		"TrySetWithTTL": 1,
		"Set":           2,
		"TrySet":        3,
		"Del":           4,
	}
)

func (x RaftLogCommand) Enum() *RaftLogCommand {
	p := new(RaftLogCommand)
	*p = x
	return p
}

func (x RaftLogCommand) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RaftLogCommand) Descriptor() protoreflect.EnumDescriptor {
	return file_node_proto_enumTypes[0].Descriptor()
}

func (RaftLogCommand) Type() protoreflect.EnumType {
	return &file_node_proto_enumTypes[0]
}

func (x RaftLogCommand) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RaftLogCommand.Descriptor instead.
func (RaftLogCommand) EnumDescriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{0}
}

type RaftState int32

const (
	RaftState_follower  RaftState = 0
	RaftState_candidate RaftState = 1
	RaftState_leader    RaftState = 2
	RaftState_shutdown  RaftState = 3
	RaftState_unknown   RaftState = 4
)

// Enum value maps for RaftState.
var (
	RaftState_name = map[int32]string{
		0: "follower",
		1: "candidate",
		2: "leader",
		3: "shutdown",
		4: "unknown",
	}
	RaftState_value = map[string]int32{
		"follower":  0,
		"candidate": 1,
		"leader":    2,
		"shutdown":  3,
		"unknown":   4,
	}
)

func (x RaftState) Enum() *RaftState {
	p := new(RaftState)
	*p = x
	return p
}

func (x RaftState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RaftState) Descriptor() protoreflect.EnumDescriptor {
	return file_node_proto_enumTypes[1].Descriptor()
}

func (RaftState) Type() protoreflect.EnumType {
	return &file_node_proto_enumTypes[1]
}

func (x RaftState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RaftState.Descriptor instead.
func (RaftState) EnumDescriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{1}
}

type RaftLogPayload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Command   RaftLogCommand `protobuf:"varint,1,opt,name=command,proto3,enum=serverpb.RaftLogCommand" json:"command,omitempty"`
	Key       []byte         `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	Value     []byte         `protobuf:"bytes,3,opt,name=value,proto3,oneof" json:"value,omitempty"`
	Ttl       *uint32        `protobuf:"varint,4,opt,name=ttl,proto3,oneof" json:"ttl,omitempty"`
	Namespace *string        `protobuf:"bytes,5,opt,name=namespace,proto3,oneof" json:"namespace,omitempty"`
}

func (x *RaftLogPayload) Reset() {
	*x = RaftLogPayload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RaftLogPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RaftLogPayload) ProtoMessage() {}

func (x *RaftLogPayload) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RaftLogPayload.ProtoReflect.Descriptor instead.
func (*RaftLogPayload) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{0}
}

func (x *RaftLogPayload) GetCommand() RaftLogCommand {
	if x != nil {
		return x.Command
	}
	return RaftLogCommand_SetWithTTL
}

func (x *RaftLogPayload) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *RaftLogPayload) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *RaftLogPayload) GetTtl() uint32 {
	if x != nil && x.Ttl != nil {
		return *x.Ttl
	}
	return 0
}

func (x *RaftLogPayload) GetNamespace() string {
	if x != nil && x.Namespace != nil {
		return *x.Namespace
	}
	return ""
}

type AppendClusterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServerId string `protobuf:"bytes,1,opt,name=server_id,json=serverId,proto3" json:"server_id,omitempty"`
	PeerAddr string `protobuf:"bytes,2,opt,name=peer_addr,json=peerAddr,proto3" json:"peer_addr,omitempty"`
	Voter    bool   `protobuf:"varint,3,opt,name=voter,proto3" json:"voter,omitempty"`
}

func (x *AppendClusterRequest) Reset() {
	*x = AppendClusterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AppendClusterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppendClusterRequest) ProtoMessage() {}

func (x *AppendClusterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppendClusterRequest.ProtoReflect.Descriptor instead.
func (*AppendClusterRequest) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{1}
}

func (x *AppendClusterRequest) GetServerId() string {
	if x != nil {
		return x.ServerId
	}
	return ""
}

func (x *AppendClusterRequest) GetPeerAddr() string {
	if x != nil {
		return x.PeerAddr
	}
	return ""
}

func (x *AppendClusterRequest) GetVoter() bool {
	if x != nil {
		return x.Voter
	}
	return false
}

type AppendClusterResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AppendClusterResponse) Reset() {
	*x = AppendClusterResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AppendClusterResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppendClusterResponse) ProtoMessage() {}

func (x *AppendClusterResponse) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppendClusterResponse.ProtoReflect.Descriptor instead.
func (*AppendClusterResponse) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{2}
}

type LeaderMonitorRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LeaderMonitorRequest) Reset() {
	*x = LeaderMonitorRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LeaderMonitorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LeaderMonitorRequest) ProtoMessage() {}

func (x *LeaderMonitorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LeaderMonitorRequest.ProtoReflect.Descriptor instead.
func (*LeaderMonitorRequest) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{3}
}

type LeaderMonitorResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Leader bool `protobuf:"varint,1,opt,name=leader,proto3" json:"leader,omitempty"`
}

func (x *LeaderMonitorResponse) Reset() {
	*x = LeaderMonitorResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LeaderMonitorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LeaderMonitorResponse) ProtoMessage() {}

func (x *LeaderMonitorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LeaderMonitorResponse.ProtoReflect.Descriptor instead.
func (*LeaderMonitorResponse) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{4}
}

func (x *LeaderMonitorResponse) GetLeader() bool {
	if x != nil {
		return x.Leader
	}
	return false
}

type RaftStateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RaftStateRequest) Reset() {
	*x = RaftStateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RaftStateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RaftStateRequest) ProtoMessage() {}

func (x *RaftStateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RaftStateRequest.ProtoReflect.Descriptor instead.
func (*RaftStateRequest) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{5}
}

type RaftStateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State RaftState `protobuf:"varint,1,opt,name=state,proto3,enum=serverpb.RaftState" json:"state,omitempty"`
}

func (x *RaftStateResponse) Reset() {
	*x = RaftStateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_node_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RaftStateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RaftStateResponse) ProtoMessage() {}

func (x *RaftStateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_node_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RaftStateResponse.ProtoReflect.Descriptor instead.
func (*RaftStateResponse) Descriptor() ([]byte, []int) {
	return file_node_proto_rawDescGZIP(), []int{6}
}

func (x *RaftStateResponse) GetState() RaftState {
	if x != nil {
		return x.State
	}
	return RaftState_follower
}

var File_node_proto protoreflect.FileDescriptor

var file_node_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x70, 0x62, 0x22, 0xcb, 0x01, 0x0a, 0x0e, 0x52, 0x61, 0x66, 0x74, 0x4c,
	0x6f, 0x67, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x32, 0x0a, 0x07, 0x63, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x18, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x52, 0x61, 0x66, 0x74, 0x4c, 0x6f, 0x67, 0x43, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x19, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x88, 0x01, 0x01, 0x12, 0x15, 0x0a, 0x03, 0x74, 0x74,
	0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x48, 0x01, 0x52, 0x03, 0x74, 0x74, 0x6c, 0x88, 0x01,
	0x01, 0x12, 0x21, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x02, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x88, 0x01, 0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x06,
	0x0a, 0x04, 0x5f, 0x74, 0x74, 0x6c, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x22, 0x66, 0x0a, 0x14, 0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x43, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x65, 0x65,
	0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x65,
	0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x6f, 0x74, 0x65, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x76, 0x6f, 0x74, 0x65, 0x72, 0x22, 0x17, 0x0a, 0x15,
	0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x16, 0x0a, 0x14, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x4d,
	0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x2f, 0x0a,
	0x15, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x22, 0x12,
	0x0a, 0x10, 0x52, 0x61, 0x66, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x22, 0x3e, 0x0a, 0x11, 0x52, 0x61, 0x66, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x29, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x70,
	0x62, 0x2e, 0x52, 0x61, 0x66, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x2a, 0x51, 0x0a, 0x0e, 0x52, 0x61, 0x66, 0x74, 0x4c, 0x6f, 0x67, 0x43, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x12, 0x0e, 0x0a, 0x0a, 0x53, 0x65, 0x74, 0x57, 0x69, 0x74, 0x68, 0x54,
	0x54, 0x4c, 0x10, 0x00, 0x12, 0x11, 0x0a, 0x0d, 0x54, 0x72, 0x79, 0x53, 0x65, 0x74, 0x57, 0x69,
	0x74, 0x68, 0x54, 0x54, 0x4c, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03, 0x53, 0x65, 0x74, 0x10, 0x02,
	0x12, 0x0a, 0x0a, 0x06, 0x54, 0x72, 0x79, 0x53, 0x65, 0x74, 0x10, 0x03, 0x12, 0x07, 0x0a, 0x03,
	0x44, 0x65, 0x6c, 0x10, 0x04, 0x2a, 0x4f, 0x0a, 0x09, 0x52, 0x61, 0x66, 0x74, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x0c, 0x0a, 0x08, 0x66, 0x6f, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x72, 0x10, 0x00,
	0x12, 0x0d, 0x0a, 0x09, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x10, 0x01, 0x12,
	0x0a, 0x0a, 0x06, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x73,
	0x68, 0x75, 0x74, 0x64, 0x6f, 0x77, 0x6e, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x75, 0x6e, 0x6b,
	0x6e, 0x6f, 0x77, 0x6e, 0x10, 0x04, 0x32, 0xfc, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x64, 0x51, 0x75,
	0x65, 0x65, 0x6e, 0x12, 0x52, 0x0a, 0x0d, 0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x12, 0x1e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x70, 0x62, 0x2e,
	0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x70, 0x62, 0x2e,
	0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x54, 0x0a, 0x0d, 0x4c, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x12, 0x1e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x70, 0x62, 0x2e, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f,
	0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x70, 0x62, 0x2e, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f,
	0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x46, 0x0a,
	0x09, 0x52, 0x61, 0x66, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1a, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x52, 0x61, 0x66, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x70,
	0x62, 0x2e, 0x52, 0x61, 0x66, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x0d, 0x5a, 0x0b, 0x2e, 0x2f, 0x3b, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_node_proto_rawDescOnce sync.Once
	file_node_proto_rawDescData = file_node_proto_rawDesc
)

func file_node_proto_rawDescGZIP() []byte {
	file_node_proto_rawDescOnce.Do(func() {
		file_node_proto_rawDescData = protoimpl.X.CompressGZIP(file_node_proto_rawDescData)
	})
	return file_node_proto_rawDescData
}

var file_node_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_node_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_node_proto_goTypes = []interface{}{
	(RaftLogCommand)(0),           // 0: serverpb.RaftLogCommand
	(RaftState)(0),                // 1: serverpb.RaftState
	(*RaftLogPayload)(nil),        // 2: serverpb.RaftLogPayload
	(*AppendClusterRequest)(nil),  // 3: serverpb.AppendClusterRequest
	(*AppendClusterResponse)(nil), // 4: serverpb.AppendClusterResponse
	(*LeaderMonitorRequest)(nil),  // 5: serverpb.LeaderMonitorRequest
	(*LeaderMonitorResponse)(nil), // 6: serverpb.LeaderMonitorResponse
	(*RaftStateRequest)(nil),      // 7: serverpb.RaftStateRequest
	(*RaftStateResponse)(nil),     // 8: serverpb.RaftStateResponse
}
var file_node_proto_depIdxs = []int32{
	0, // 0: serverpb.RaftLogPayload.command:type_name -> serverpb.RaftLogCommand
	1, // 1: serverpb.RaftStateResponse.state:type_name -> serverpb.RaftState
	3, // 2: serverpb.RedQueen.AppendCluster:input_type -> serverpb.AppendClusterRequest
	5, // 3: serverpb.RedQueen.LeaderMonitor:input_type -> serverpb.LeaderMonitorRequest
	7, // 4: serverpb.RedQueen.RaftState:input_type -> serverpb.RaftStateRequest
	4, // 5: serverpb.RedQueen.AppendCluster:output_type -> serverpb.AppendClusterResponse
	6, // 6: serverpb.RedQueen.LeaderMonitor:output_type -> serverpb.LeaderMonitorResponse
	8, // 7: serverpb.RedQueen.RaftState:output_type -> serverpb.RaftStateResponse
	5, // [5:8] is the sub-list for method output_type
	2, // [2:5] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_node_proto_init() }
func file_node_proto_init() {
	if File_node_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_node_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RaftLogPayload); i {
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
		file_node_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AppendClusterRequest); i {
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
		file_node_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AppendClusterResponse); i {
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
		file_node_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LeaderMonitorRequest); i {
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
		file_node_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LeaderMonitorResponse); i {
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
		file_node_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RaftStateRequest); i {
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
		file_node_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RaftStateResponse); i {
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
	file_node_proto_msgTypes[0].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_node_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_node_proto_goTypes,
		DependencyIndexes: file_node_proto_depIdxs,
		EnumInfos:         file_node_proto_enumTypes,
		MessageInfos:      file_node_proto_msgTypes,
	}.Build()
	File_node_proto = out.File
	file_node_proto_rawDesc = nil
	file_node_proto_goTypes = nil
	file_node_proto_depIdxs = nil
}
