// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: language.proto

package language

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

// Api Endpoints for LanguageService service

func NewLanguageServiceEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for LanguageService service

type LanguageService interface {
	// 查找当前domain下的所有多语言(目前中英日三种)
	FindLanguages(ctx context.Context, in *FindLanguagesRequest, opts ...client.CallOption) (*FindLanguagesResponse, error)
	// 查找当前domain和langcd下面的语言
	FindLanguage(ctx context.Context, in *FindLanguageRequest, opts ...client.CallOption) (*FindLanguageResponse, error)
	// 通过当前domain、langcd和对应的key，获取下面的语言结果
	FindLanguageValue(ctx context.Context, in *FindLanguageValueRequest, opts ...client.CallOption) (*FindLanguageValueResponse, error)
	// 添加一种多语言数据
	AddLanguage(ctx context.Context, in *AddLanguageRequest, opts ...client.CallOption) (*AddLanguageResponse, error)
	// 添加某个公司-某个APP-某种多语言的相关多语言数据(domain + app_id + langcd)
	AddLanguageData(ctx context.Context, in *AddLanguageDataRequest, opts ...client.CallOption) (*Response, error)
	// 添加某个公司-某个APP-某个子项目集-某个子项目-某种多语言的相关多语言数据(domain + app_id + type + key + langcd)
	AddAppLanguageData(ctx context.Context, in *AddAppLanguageDataRequest, opts ...client.CallOption) (*Response, error)
	// 添加某个公司共通项目的多语言数据(domain + type + key + langcd)
	AddCommonData(ctx context.Context, in *AddCommonDataRequest, opts ...client.CallOption) (*Response, error)
	// 添加或更新多条多语言数据(domain + app_id + type + key + langcd)
	AddManyLanData(ctx context.Context, in *AddManyLanDataRequest, opts ...client.CallOption) (*Response, error)
	// 删除某个公司共通项目的多语言数据(domain + type + key + langcd)
	DeleteCommonData(ctx context.Context, in *DeleteCommonDataRequest, opts ...client.CallOption) (*Response, error)
	// 删除某个公司-某个APP-某种多语言的相关多语言数据(domain + app_id)
	DeleteLanguageData(ctx context.Context, in *DeleteLanguageDataRequest, opts ...client.CallOption) (*Response, error)
	// 删除添加某个公司-某个APP-某个子项目集-某个子项目的相关多语言数据(domain + app_id + type + key)
	DeleteAppLanguageData(ctx context.Context, in *DeleteAppLanguageDataRequest, opts ...client.CallOption) (*Response, error)
}

type languageService struct {
	c    client.Client
	name string
}

func NewLanguageService(name string, c client.Client) LanguageService {
	return &languageService{
		c:    c,
		name: name,
	}
}

func (c *languageService) FindLanguages(ctx context.Context, in *FindLanguagesRequest, opts ...client.CallOption) (*FindLanguagesResponse, error) {
	req := c.c.NewRequest(c.name, "LanguageService.FindLanguages", in)
	out := new(FindLanguagesResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) FindLanguage(ctx context.Context, in *FindLanguageRequest, opts ...client.CallOption) (*FindLanguageResponse, error) {
	req := c.c.NewRequest(c.name, "LanguageService.FindLanguage", in)
	out := new(FindLanguageResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) FindLanguageValue(ctx context.Context, in *FindLanguageValueRequest, opts ...client.CallOption) (*FindLanguageValueResponse, error) {
	req := c.c.NewRequest(c.name, "LanguageService.FindLanguageValue", in)
	out := new(FindLanguageValueResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) AddLanguage(ctx context.Context, in *AddLanguageRequest, opts ...client.CallOption) (*AddLanguageResponse, error) {
	req := c.c.NewRequest(c.name, "LanguageService.AddLanguage", in)
	out := new(AddLanguageResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) AddLanguageData(ctx context.Context, in *AddLanguageDataRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "LanguageService.AddLanguageData", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) AddAppLanguageData(ctx context.Context, in *AddAppLanguageDataRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "LanguageService.AddAppLanguageData", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) AddCommonData(ctx context.Context, in *AddCommonDataRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "LanguageService.AddCommonData", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) AddManyLanData(ctx context.Context, in *AddManyLanDataRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "LanguageService.AddManyLanData", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) DeleteCommonData(ctx context.Context, in *DeleteCommonDataRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "LanguageService.DeleteCommonData", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) DeleteLanguageData(ctx context.Context, in *DeleteLanguageDataRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "LanguageService.DeleteLanguageData", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *languageService) DeleteAppLanguageData(ctx context.Context, in *DeleteAppLanguageDataRequest, opts ...client.CallOption) (*Response, error) {
	req := c.c.NewRequest(c.name, "LanguageService.DeleteAppLanguageData", in)
	out := new(Response)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for LanguageService service

type LanguageServiceHandler interface {
	// 查找当前domain下的所有多语言(目前中英日三种)
	FindLanguages(context.Context, *FindLanguagesRequest, *FindLanguagesResponse) error
	// 查找当前domain和langcd下面的语言
	FindLanguage(context.Context, *FindLanguageRequest, *FindLanguageResponse) error
	// 通过当前domain、langcd和对应的key，获取下面的语言结果
	FindLanguageValue(context.Context, *FindLanguageValueRequest, *FindLanguageValueResponse) error
	// 添加一种多语言数据
	AddLanguage(context.Context, *AddLanguageRequest, *AddLanguageResponse) error
	// 添加某个公司-某个APP-某种多语言的相关多语言数据(domain + app_id + langcd)
	AddLanguageData(context.Context, *AddLanguageDataRequest, *Response) error
	// 添加某个公司-某个APP-某个子项目集-某个子项目-某种多语言的相关多语言数据(domain + app_id + type + key + langcd)
	AddAppLanguageData(context.Context, *AddAppLanguageDataRequest, *Response) error
	// 添加某个公司共通项目的多语言数据(domain + type + key + langcd)
	AddCommonData(context.Context, *AddCommonDataRequest, *Response) error
	// 添加或更新多条多语言数据(domain + app_id + type + key + langcd)
	AddManyLanData(context.Context, *AddManyLanDataRequest, *Response) error
	// 删除某个公司共通项目的多语言数据(domain + type + key + langcd)
	DeleteCommonData(context.Context, *DeleteCommonDataRequest, *Response) error
	// 删除某个公司-某个APP-某种多语言的相关多语言数据(domain + app_id)
	DeleteLanguageData(context.Context, *DeleteLanguageDataRequest, *Response) error
	// 删除添加某个公司-某个APP-某个子项目集-某个子项目的相关多语言数据(domain + app_id + type + key)
	DeleteAppLanguageData(context.Context, *DeleteAppLanguageDataRequest, *Response) error
}

func RegisterLanguageServiceHandler(s server.Server, hdlr LanguageServiceHandler, opts ...server.HandlerOption) error {
	type languageService interface {
		FindLanguages(ctx context.Context, in *FindLanguagesRequest, out *FindLanguagesResponse) error
		FindLanguage(ctx context.Context, in *FindLanguageRequest, out *FindLanguageResponse) error
		FindLanguageValue(ctx context.Context, in *FindLanguageValueRequest, out *FindLanguageValueResponse) error
		AddLanguage(ctx context.Context, in *AddLanguageRequest, out *AddLanguageResponse) error
		AddLanguageData(ctx context.Context, in *AddLanguageDataRequest, out *Response) error
		AddAppLanguageData(ctx context.Context, in *AddAppLanguageDataRequest, out *Response) error
		AddCommonData(ctx context.Context, in *AddCommonDataRequest, out *Response) error
		AddManyLanData(ctx context.Context, in *AddManyLanDataRequest, out *Response) error
		DeleteCommonData(ctx context.Context, in *DeleteCommonDataRequest, out *Response) error
		DeleteLanguageData(ctx context.Context, in *DeleteLanguageDataRequest, out *Response) error
		DeleteAppLanguageData(ctx context.Context, in *DeleteAppLanguageDataRequest, out *Response) error
	}
	type LanguageService struct {
		languageService
	}
	h := &languageServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&LanguageService{h}, opts...))
}

type languageServiceHandler struct {
	LanguageServiceHandler
}

func (h *languageServiceHandler) FindLanguages(ctx context.Context, in *FindLanguagesRequest, out *FindLanguagesResponse) error {
	return h.LanguageServiceHandler.FindLanguages(ctx, in, out)
}

func (h *languageServiceHandler) FindLanguage(ctx context.Context, in *FindLanguageRequest, out *FindLanguageResponse) error {
	return h.LanguageServiceHandler.FindLanguage(ctx, in, out)
}

func (h *languageServiceHandler) FindLanguageValue(ctx context.Context, in *FindLanguageValueRequest, out *FindLanguageValueResponse) error {
	return h.LanguageServiceHandler.FindLanguageValue(ctx, in, out)
}

func (h *languageServiceHandler) AddLanguage(ctx context.Context, in *AddLanguageRequest, out *AddLanguageResponse) error {
	return h.LanguageServiceHandler.AddLanguage(ctx, in, out)
}

func (h *languageServiceHandler) AddLanguageData(ctx context.Context, in *AddLanguageDataRequest, out *Response) error {
	return h.LanguageServiceHandler.AddLanguageData(ctx, in, out)
}

func (h *languageServiceHandler) AddAppLanguageData(ctx context.Context, in *AddAppLanguageDataRequest, out *Response) error {
	return h.LanguageServiceHandler.AddAppLanguageData(ctx, in, out)
}

func (h *languageServiceHandler) AddCommonData(ctx context.Context, in *AddCommonDataRequest, out *Response) error {
	return h.LanguageServiceHandler.AddCommonData(ctx, in, out)
}

func (h *languageServiceHandler) AddManyLanData(ctx context.Context, in *AddManyLanDataRequest, out *Response) error {
	return h.LanguageServiceHandler.AddManyLanData(ctx, in, out)
}

func (h *languageServiceHandler) DeleteCommonData(ctx context.Context, in *DeleteCommonDataRequest, out *Response) error {
	return h.LanguageServiceHandler.DeleteCommonData(ctx, in, out)
}

func (h *languageServiceHandler) DeleteLanguageData(ctx context.Context, in *DeleteLanguageDataRequest, out *Response) error {
	return h.LanguageServiceHandler.DeleteLanguageData(ctx, in, out)
}

func (h *languageServiceHandler) DeleteAppLanguageData(ctx context.Context, in *DeleteAppLanguageDataRequest, out *Response) error {
	return h.LanguageServiceHandler.DeleteAppLanguageData(ctx, in, out)
}
