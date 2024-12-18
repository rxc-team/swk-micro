// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: check.proto

package check

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

import (
	context "context"
	api "github.com/micro/go-micro/v2/api"
	client "github.com/micro/go-micro/v2/client"
	server "github.com/micro/go-micro/v2/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option

// Api Endpoints for CheckHistoryService service

func NewCheckHistoryServiceEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for CheckHistoryService service

type CheckHistoryService interface {
	FindHistories(ctx context.Context, in *HistoriesRequest, opts ...client.CallOption) (*HistoriesResponse, error)
	FindHistoryCount(ctx context.Context, in *CountRequest, opts ...client.CallOption) (*CountResponse, error)
	DeleteHistories(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error)
	Download(ctx context.Context, in *DownloadRequest, opts ...client.CallOption) (CheckHistoryService_DownloadService, error)
}

type checkHistoryService struct {
	c    client.Client
	name string
}

func NewCheckHistoryService(name string, c client.Client) CheckHistoryService {
	return &checkHistoryService{
		c:    c,
		name: name,
	}
}

func (c *checkHistoryService) FindHistories(ctx context.Context, in *HistoriesRequest, opts ...client.CallOption) (*HistoriesResponse, error) {
	req := c.c.NewRequest(c.name, "CheckHistoryService.FindHistories", in)
	out := new(HistoriesResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *checkHistoryService) FindHistoryCount(ctx context.Context, in *CountRequest, opts ...client.CallOption) (*CountResponse, error) {
	req := c.c.NewRequest(c.name, "CheckHistoryService.FindHistoryCount", in)
	out := new(CountResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *checkHistoryService) DeleteHistories(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error) {
	req := c.c.NewRequest(c.name, "CheckHistoryService.DeleteHistories", in)
	out := new(DeleteResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *checkHistoryService) Download(ctx context.Context, in *DownloadRequest, opts ...client.CallOption) (CheckHistoryService_DownloadService, error) {
	req := c.c.NewRequest(c.name, "CheckHistoryService.Download", &DownloadRequest{})
	stream, err := c.c.Stream(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(in); err != nil {
		return nil, err
	}
	return &checkHistoryServiceDownload{stream}, nil
}

type CheckHistoryService_DownloadService interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	Close() error
	Recv() (*DownloadResponse, error)
}

type checkHistoryServiceDownload struct {
	stream client.Stream
}

func (x *checkHistoryServiceDownload) Close() error {
	return x.stream.Close()
}

func (x *checkHistoryServiceDownload) Context() context.Context {
	return x.stream.Context()
}

func (x *checkHistoryServiceDownload) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *checkHistoryServiceDownload) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *checkHistoryServiceDownload) Recv() (*DownloadResponse, error) {
	m := new(DownloadResponse)
	err := x.stream.Recv(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for CheckHistoryService service

type CheckHistoryServiceHandler interface {
	FindHistories(context.Context, *HistoriesRequest, *HistoriesResponse) error
	FindHistoryCount(context.Context, *CountRequest, *CountResponse) error
	DeleteHistories(context.Context, *DeleteRequest, *DeleteResponse) error
	Download(context.Context, *DownloadRequest, CheckHistoryService_DownloadStream) error
}

func RegisterCheckHistoryServiceHandler(s server.Server, hdlr CheckHistoryServiceHandler, opts ...server.HandlerOption) error {
	type checkHistoryService interface {
		FindHistories(ctx context.Context, in *HistoriesRequest, out *HistoriesResponse) error
		FindHistoryCount(ctx context.Context, in *CountRequest, out *CountResponse) error
		DeleteHistories(ctx context.Context, in *DeleteRequest, out *DeleteResponse) error
		Download(ctx context.Context, stream server.Stream) error
	}
	type CheckHistoryService struct {
		checkHistoryService
	}
	h := &checkHistoryServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&CheckHistoryService{h}, opts...))
}

type checkHistoryServiceHandler struct {
	CheckHistoryServiceHandler
}

func (h *checkHistoryServiceHandler) FindHistories(ctx context.Context, in *HistoriesRequest, out *HistoriesResponse) error {
	return h.CheckHistoryServiceHandler.FindHistories(ctx, in, out)
}

func (h *checkHistoryServiceHandler) FindHistoryCount(ctx context.Context, in *CountRequest, out *CountResponse) error {
	return h.CheckHistoryServiceHandler.FindHistoryCount(ctx, in, out)
}

func (h *checkHistoryServiceHandler) DeleteHistories(ctx context.Context, in *DeleteRequest, out *DeleteResponse) error {
	return h.CheckHistoryServiceHandler.DeleteHistories(ctx, in, out)
}

func (h *checkHistoryServiceHandler) Download(ctx context.Context, stream server.Stream) error {
	m := new(DownloadRequest)
	if err := stream.Recv(m); err != nil {
		return err
	}
	return h.CheckHistoryServiceHandler.Download(ctx, m, &checkHistoryServiceDownloadStream{stream})
}

type CheckHistoryService_DownloadStream interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	Close() error
	Send(*DownloadResponse) error
}

type checkHistoryServiceDownloadStream struct {
	stream server.Stream
}

func (x *checkHistoryServiceDownloadStream) Close() error {
	return x.stream.Close()
}

func (x *checkHistoryServiceDownloadStream) Context() context.Context {
	return x.stream.Context()
}

func (x *checkHistoryServiceDownloadStream) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *checkHistoryServiceDownloadStream) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *checkHistoryServiceDownloadStream) Send(m *DownloadResponse) error {
	return x.stream.Send(m)
}
