// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.23.4
// source: node.proto

package serverpb

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

// RedQueenClient is the client API for RedQueen service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RedQueenClient interface {
	AppendCluster(ctx context.Context, in *AppendClusterRequest, opts ...grpc.CallOption) (*AppendClusterResponse, error)
	LeaderMonitor(ctx context.Context, in *LeaderMonitorRequest, opts ...grpc.CallOption) (RedQueen_LeaderMonitorClient, error)
	RaftState(ctx context.Context, in *RaftStateRequest, opts ...grpc.CallOption) (*RaftStateResponse, error)
}

type redQueenClient struct {
	cc grpc.ClientConnInterface
}

func NewRedQueenClient(cc grpc.ClientConnInterface) RedQueenClient {
	return &redQueenClient{cc}
}

func (c *redQueenClient) AppendCluster(ctx context.Context, in *AppendClusterRequest, opts ...grpc.CallOption) (*AppendClusterResponse, error) {
	out := new(AppendClusterResponse)
	err := c.cc.Invoke(ctx, "/serverpb.RedQueen/AppendCluster", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *redQueenClient) LeaderMonitor(ctx context.Context, in *LeaderMonitorRequest, opts ...grpc.CallOption) (RedQueen_LeaderMonitorClient, error) {
	stream, err := c.cc.NewStream(ctx, &RedQueen_ServiceDesc.Streams[0], "/serverpb.RedQueen/LeaderMonitor", opts...)
	if err != nil {
		return nil, err
	}
	x := &redQueenLeaderMonitorClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type RedQueen_LeaderMonitorClient interface {
	Recv() (*LeaderMonitorResponse, error)
	grpc.ClientStream
}

type redQueenLeaderMonitorClient struct {
	grpc.ClientStream
}

func (x *redQueenLeaderMonitorClient) Recv() (*LeaderMonitorResponse, error) {
	m := new(LeaderMonitorResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *redQueenClient) RaftState(ctx context.Context, in *RaftStateRequest, opts ...grpc.CallOption) (*RaftStateResponse, error) {
	out := new(RaftStateResponse)
	err := c.cc.Invoke(ctx, "/serverpb.RedQueen/RaftState", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RedQueenServer is the server API for RedQueen service.
// All implementations must embed UnimplementedRedQueenServer
// for forward compatibility
type RedQueenServer interface {
	AppendCluster(context.Context, *AppendClusterRequest) (*AppendClusterResponse, error)
	LeaderMonitor(*LeaderMonitorRequest, RedQueen_LeaderMonitorServer) error
	RaftState(context.Context, *RaftStateRequest) (*RaftStateResponse, error)
	mustEmbedUnimplementedRedQueenServer()
}

// UnimplementedRedQueenServer must be embedded to have forward compatible implementations.
type UnimplementedRedQueenServer struct {
}

func (UnimplementedRedQueenServer) AppendCluster(context.Context, *AppendClusterRequest) (*AppendClusterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendCluster not implemented")
}
func (UnimplementedRedQueenServer) LeaderMonitor(*LeaderMonitorRequest, RedQueen_LeaderMonitorServer) error {
	return status.Errorf(codes.Unimplemented, "method LeaderMonitor not implemented")
}
func (UnimplementedRedQueenServer) RaftState(context.Context, *RaftStateRequest) (*RaftStateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RaftState not implemented")
}
func (UnimplementedRedQueenServer) mustEmbedUnimplementedRedQueenServer() {}

// UnsafeRedQueenServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RedQueenServer will
// result in compilation errors.
type UnsafeRedQueenServer interface {
	mustEmbedUnimplementedRedQueenServer()
}

func RegisterRedQueenServer(s grpc.ServiceRegistrar, srv RedQueenServer) {
	s.RegisterService(&RedQueen_ServiceDesc, srv)
}

func _RedQueen_AppendCluster_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppendClusterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RedQueenServer).AppendCluster(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverpb.RedQueen/AppendCluster",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RedQueenServer).AppendCluster(ctx, req.(*AppendClusterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RedQueen_LeaderMonitor_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LeaderMonitorRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RedQueenServer).LeaderMonitor(m, &redQueenLeaderMonitorServer{stream})
}

type RedQueen_LeaderMonitorServer interface {
	Send(*LeaderMonitorResponse) error
	grpc.ServerStream
}

type redQueenLeaderMonitorServer struct {
	grpc.ServerStream
}

func (x *redQueenLeaderMonitorServer) Send(m *LeaderMonitorResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _RedQueen_RaftState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RaftStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RedQueenServer).RaftState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverpb.RedQueen/RaftState",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RedQueenServer).RaftState(ctx, req.(*RaftStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RedQueen_ServiceDesc is the grpc.ServiceDesc for RedQueen service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RedQueen_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "serverpb.RedQueen",
	HandlerType: (*RedQueenServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AppendCluster",
			Handler:    _RedQueen_AppendCluster_Handler,
		},
		{
			MethodName: "RaftState",
			Handler:    _RedQueen_RaftState_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "LeaderMonitor",
			Handler:       _RedQueen_LeaderMonitor_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "node.proto",
}
