// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: relation.proto

package relation

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

// Api Endpoints for RelationService service

func NewRelationServiceEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for RelationService service

type RelationService interface {
	FindRelations(ctx context.Context, in *RelationsRequest, opts ...client.CallOption) (*RelationsResponse, error)
	AddRelation(ctx context.Context, in *AddRequest, opts ...client.CallOption) (*AddResponse, error)
	DeleteRelation(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error)
}

type relationService struct {
	c    client.Client
	name string
}

func NewRelationService(name string, c client.Client) RelationService {
	return &relationService{
		c:    c,
		name: name,
	}
}

func (c *relationService) FindRelations(ctx context.Context, in *RelationsRequest, opts ...client.CallOption) (*RelationsResponse, error) {
	req := c.c.NewRequest(c.name, "RelationService.FindRelations", in)
	out := new(RelationsResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *relationService) AddRelation(ctx context.Context, in *AddRequest, opts ...client.CallOption) (*AddResponse, error) {
	req := c.c.NewRequest(c.name, "RelationService.AddRelation", in)
	out := new(AddResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *relationService) DeleteRelation(ctx context.Context, in *DeleteRequest, opts ...client.CallOption) (*DeleteResponse, error) {
	req := c.c.NewRequest(c.name, "RelationService.DeleteRelation", in)
	out := new(DeleteResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for RelationService service

type RelationServiceHandler interface {
	FindRelations(context.Context, *RelationsRequest, *RelationsResponse) error
	AddRelation(context.Context, *AddRequest, *AddResponse) error
	DeleteRelation(context.Context, *DeleteRequest, *DeleteResponse) error
}

func RegisterRelationServiceHandler(s server.Server, hdlr RelationServiceHandler, opts ...server.HandlerOption) error {
	type relationService interface {
		FindRelations(ctx context.Context, in *RelationsRequest, out *RelationsResponse) error
		AddRelation(ctx context.Context, in *AddRequest, out *AddResponse) error
		DeleteRelation(ctx context.Context, in *DeleteRequest, out *DeleteResponse) error
	}
	type RelationService struct {
		relationService
	}
	h := &relationServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&RelationService{h}, opts...))
}

type relationServiceHandler struct {
	RelationServiceHandler
}

func (h *relationServiceHandler) FindRelations(ctx context.Context, in *RelationsRequest, out *RelationsResponse) error {
	return h.RelationServiceHandler.FindRelations(ctx, in, out)
}

func (h *relationServiceHandler) AddRelation(ctx context.Context, in *AddRequest, out *AddResponse) error {
	return h.RelationServiceHandler.AddRelation(ctx, in, out)
}

func (h *relationServiceHandler) DeleteRelation(ctx context.Context, in *DeleteRequest, out *DeleteResponse) error {
	return h.RelationServiceHandler.DeleteRelation(ctx, in, out)
}
