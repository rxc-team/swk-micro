// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: datapatch.proto

package datapatch

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

// Api Endpoints for DataPatchService service

func NewDataPatchServiceEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for DataPatchService service

type DataPatchService interface {
	DataPatch1216(ctx context.Context, in *DataPatch1216Request, opts ...client.CallOption) (*DataPatch1216Response, error)
}

type dataPatchService struct {
	c    client.Client
	name string
}

func NewDataPatchService(name string, c client.Client) DataPatchService {
	return &dataPatchService{
		c:    c,
		name: name,
	}
}

func (c *dataPatchService) DataPatch1216(ctx context.Context, in *DataPatch1216Request, opts ...client.CallOption) (*DataPatch1216Response, error) {
	req := c.c.NewRequest(c.name, "DataPatchService.DataPatch1216", in)
	out := new(DataPatch1216Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for DataPatchService service

type DataPatchServiceHandler interface {
	DataPatch1216(context.Context, *DataPatch1216Request, *DataPatch1216Response) error
}

func RegisterDataPatchServiceHandler(s server.Server, hdlr DataPatchServiceHandler, opts ...server.HandlerOption) error {
	type dataPatchService interface {
		DataPatch1216(ctx context.Context, in *DataPatch1216Request, out *DataPatch1216Response) error
	}
	type DataPatchService struct {
		dataPatchService
	}
	h := &dataPatchServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&DataPatchService{h}, opts...))
}

type dataPatchServiceHandler struct {
	DataPatchServiceHandler
}

func (h *dataPatchServiceHandler) DataPatch1216(ctx context.Context, in *DataPatch1216Request, out *DataPatch1216Response) error {
	return h.DataPatchServiceHandler.DataPatch1216(ctx, in, out)
}
