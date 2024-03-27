// Code generated by protoc-gen-go-drpc. DO NOT EDIT.
// protoc-gen-go-drpc version: v0.0.33
// source: insightd/proto/insightd.proto

package proto

import (
	context "context"
	errors "errors"
	protojson "google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
	drpc "storj.io/drpc"
	drpcerr "storj.io/drpc/drpcerr"
)

type drpcEncoding_File_insightd_proto_insightd_proto struct{}

func (drpcEncoding_File_insightd_proto_insightd_proto) Marshal(msg drpc.Message) ([]byte, error) {
	return proto.Marshal(msg.(proto.Message))
}

func (drpcEncoding_File_insightd_proto_insightd_proto) MarshalAppend(buf []byte, msg drpc.Message) ([]byte, error) {
	return proto.MarshalOptions{}.MarshalAppend(buf, msg.(proto.Message))
}

func (drpcEncoding_File_insightd_proto_insightd_proto) Unmarshal(buf []byte, msg drpc.Message) error {
	return proto.Unmarshal(buf, msg.(proto.Message))
}

func (drpcEncoding_File_insightd_proto_insightd_proto) JSONMarshal(msg drpc.Message) ([]byte, error) {
	return protojson.Marshal(msg.(proto.Message))
}

func (drpcEncoding_File_insightd_proto_insightd_proto) JSONUnmarshal(buf []byte, msg drpc.Message) error {
	return protojson.Unmarshal(buf, msg.(proto.Message))
}

type DRPCInsightDaemonClient interface {
	DRPCConn() drpc.Conn

	Register(ctx context.Context, in *RegisterRequest) (DRPCInsightDaemon_RegisterClient, error)
	RecordInvocation(ctx context.Context, in *ReportInvocationRequest) (*Empty, error)
	ReportPath(ctx context.Context, in *ReportPathRequest) (*Empty, error)
}

type drpcInsightDaemonClient struct {
	cc drpc.Conn
}

func NewDRPCInsightDaemonClient(cc drpc.Conn) DRPCInsightDaemonClient {
	return &drpcInsightDaemonClient{cc}
}

func (c *drpcInsightDaemonClient) DRPCConn() drpc.Conn { return c.cc }

func (c *drpcInsightDaemonClient) Register(ctx context.Context, in *RegisterRequest) (DRPCInsightDaemon_RegisterClient, error) {
	stream, err := c.cc.NewStream(ctx, "/insightd.InsightDaemon/Register", drpcEncoding_File_insightd_proto_insightd_proto{})
	if err != nil {
		return nil, err
	}
	x := &drpcInsightDaemon_RegisterClient{stream}
	if err := x.MsgSend(in, drpcEncoding_File_insightd_proto_insightd_proto{}); err != nil {
		return nil, err
	}
	if err := x.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DRPCInsightDaemon_RegisterClient interface {
	drpc.Stream
	Recv() (*SystemResponse, error)
}

type drpcInsightDaemon_RegisterClient struct {
	drpc.Stream
}

func (x *drpcInsightDaemon_RegisterClient) GetStream() drpc.Stream {
	return x.Stream
}

func (x *drpcInsightDaemon_RegisterClient) Recv() (*SystemResponse, error) {
	m := new(SystemResponse)
	if err := x.MsgRecv(m, drpcEncoding_File_insightd_proto_insightd_proto{}); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *drpcInsightDaemon_RegisterClient) RecvMsg(m *SystemResponse) error {
	return x.MsgRecv(m, drpcEncoding_File_insightd_proto_insightd_proto{})
}

func (c *drpcInsightDaemonClient) RecordInvocation(ctx context.Context, in *ReportInvocationRequest) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/insightd.InsightDaemon/RecordInvocation", drpcEncoding_File_insightd_proto_insightd_proto{}, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *drpcInsightDaemonClient) ReportPath(ctx context.Context, in *ReportPathRequest) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/insightd.InsightDaemon/ReportPath", drpcEncoding_File_insightd_proto_insightd_proto{}, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type DRPCInsightDaemonServer interface {
	Register(*RegisterRequest, DRPCInsightDaemon_RegisterStream) error
	RecordInvocation(context.Context, *ReportInvocationRequest) (*Empty, error)
	ReportPath(context.Context, *ReportPathRequest) (*Empty, error)
}

type DRPCInsightDaemonUnimplementedServer struct{}

func (s *DRPCInsightDaemonUnimplementedServer) Register(*RegisterRequest, DRPCInsightDaemon_RegisterStream) error {
	return drpcerr.WithCode(errors.New("Unimplemented"), drpcerr.Unimplemented)
}

func (s *DRPCInsightDaemonUnimplementedServer) RecordInvocation(context.Context, *ReportInvocationRequest) (*Empty, error) {
	return nil, drpcerr.WithCode(errors.New("Unimplemented"), drpcerr.Unimplemented)
}

func (s *DRPCInsightDaemonUnimplementedServer) ReportPath(context.Context, *ReportPathRequest) (*Empty, error) {
	return nil, drpcerr.WithCode(errors.New("Unimplemented"), drpcerr.Unimplemented)
}

type DRPCInsightDaemonDescription struct{}

func (DRPCInsightDaemonDescription) NumMethods() int { return 3 }

func (DRPCInsightDaemonDescription) Method(n int) (string, drpc.Encoding, drpc.Receiver, interface{}, bool) {
	switch n {
	case 0:
		return "/insightd.InsightDaemon/Register", drpcEncoding_File_insightd_proto_insightd_proto{},
			func(srv interface{}, ctx context.Context, in1, in2 interface{}) (drpc.Message, error) {
				return nil, srv.(DRPCInsightDaemonServer).
					Register(
						in1.(*RegisterRequest),
						&drpcInsightDaemon_RegisterStream{in2.(drpc.Stream)},
					)
			}, DRPCInsightDaemonServer.Register, true
	case 1:
		return "/insightd.InsightDaemon/RecordInvocation", drpcEncoding_File_insightd_proto_insightd_proto{},
			func(srv interface{}, ctx context.Context, in1, in2 interface{}) (drpc.Message, error) {
				return srv.(DRPCInsightDaemonServer).
					RecordInvocation(
						ctx,
						in1.(*ReportInvocationRequest),
					)
			}, DRPCInsightDaemonServer.RecordInvocation, true
	case 2:
		return "/insightd.InsightDaemon/ReportPath", drpcEncoding_File_insightd_proto_insightd_proto{},
			func(srv interface{}, ctx context.Context, in1, in2 interface{}) (drpc.Message, error) {
				return srv.(DRPCInsightDaemonServer).
					ReportPath(
						ctx,
						in1.(*ReportPathRequest),
					)
			}, DRPCInsightDaemonServer.ReportPath, true
	default:
		return "", nil, nil, nil, false
	}
}

func DRPCRegisterInsightDaemon(mux drpc.Mux, impl DRPCInsightDaemonServer) error {
	return mux.Register(impl, DRPCInsightDaemonDescription{})
}

type DRPCInsightDaemon_RegisterStream interface {
	drpc.Stream
	Send(*SystemResponse) error
}

type drpcInsightDaemon_RegisterStream struct {
	drpc.Stream
}

func (x *drpcInsightDaemon_RegisterStream) Send(m *SystemResponse) error {
	return x.MsgSend(m, drpcEncoding_File_insightd_proto_insightd_proto{})
}

type DRPCInsightDaemon_RecordInvocationStream interface {
	drpc.Stream
	SendAndClose(*Empty) error
}

type drpcInsightDaemon_RecordInvocationStream struct {
	drpc.Stream
}

func (x *drpcInsightDaemon_RecordInvocationStream) SendAndClose(m *Empty) error {
	if err := x.MsgSend(m, drpcEncoding_File_insightd_proto_insightd_proto{}); err != nil {
		return err
	}
	return x.CloseSend()
}

type DRPCInsightDaemon_ReportPathStream interface {
	drpc.Stream
	SendAndClose(*Empty) error
}

type drpcInsightDaemon_ReportPathStream struct {
	drpc.Stream
}

func (x *drpcInsightDaemon_ReportPathStream) SendAndClose(m *Empty) error {
	if err := x.MsgSend(m, drpcEncoding_File_insightd_proto_insightd_proto{}); err != nil {
		return err
	}
	return x.CloseSend()
}