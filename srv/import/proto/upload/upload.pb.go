// Code generated by protoc-gen-go. DO NOT EDIT.
// source: upload.proto

package upload

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
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

// 基础参数
type Params struct {
	JobId                string   `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	Action               string   `protobuf:"bytes,2,opt,name=action,proto3" json:"action"`
	Encoding             string   `protobuf:"bytes,3,opt,name=encoding,proto3" json:"encoding"`
	ZipCharset           string   `protobuf:"bytes,16,opt,name=zip_charset,json=zipCharset,proto3" json:"zip_charset"`
	UserId               string   `protobuf:"bytes,4,opt,name=user_id,json=userId,proto3" json:"user_id"`
	AppId                string   `protobuf:"bytes,5,opt,name=app_id,json=appId,proto3" json:"app_id"`
	Lang                 string   `protobuf:"bytes,6,opt,name=lang,proto3" json:"lang"`
	Domain               string   `protobuf:"bytes,7,opt,name=domain,proto3" json:"domain"`
	DatastoreId          string   `protobuf:"bytes,8,opt,name=datastore_id,json=datastoreId,proto3" json:"datastore_id"`
	GroupId              string   `protobuf:"bytes,11,opt,name=group_id,json=groupId,proto3" json:"group_id"`
	EmptyChange          bool     `protobuf:"varint,17,opt,name=empty_change,json=emptyChange,proto3" json:"empty_change"`
	AccessKeys           []string `protobuf:"bytes,12,rep,name=access_keys,json=accessKeys,proto3" json:"access_keys"`
	Owners               []string `protobuf:"bytes,13,rep,name=owners,proto3" json:"owners"`
	Roles                []string `protobuf:"bytes,14,rep,name=roles,proto3" json:"roles"`
	WfId                 string   `protobuf:"bytes,15,opt,name=wf_id,json=wfId,proto3" json:"wf_id"`
	Database             string   `protobuf:"bytes,10,opt,name=database,proto3" json:"database"`
	FirstMonth           string   `protobuf:"bytes,18,opt,name=firstMonth,proto3" json:"firstMonth"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_91b94b655bd2a7e5, []int{0}
}

func (m *Params) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Params.Unmarshal(m, b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Params.Marshal(b, m, deterministic)
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return xxx_messageInfo_Params.Size(m)
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetJobId() string {
	if m != nil {
		return m.JobId
	}
	return ""
}

func (m *Params) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

func (m *Params) GetEncoding() string {
	if m != nil {
		return m.Encoding
	}
	return ""
}

func (m *Params) GetZipCharset() string {
	if m != nil {
		return m.ZipCharset
	}
	return ""
}

func (m *Params) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *Params) GetAppId() string {
	if m != nil {
		return m.AppId
	}
	return ""
}

func (m *Params) GetLang() string {
	if m != nil {
		return m.Lang
	}
	return ""
}

func (m *Params) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

func (m *Params) GetDatastoreId() string {
	if m != nil {
		return m.DatastoreId
	}
	return ""
}

func (m *Params) GetGroupId() string {
	if m != nil {
		return m.GroupId
	}
	return ""
}

func (m *Params) GetEmptyChange() bool {
	if m != nil {
		return m.EmptyChange
	}
	return false
}

func (m *Params) GetAccessKeys() []string {
	if m != nil {
		return m.AccessKeys
	}
	return nil
}

func (m *Params) GetOwners() []string {
	if m != nil {
		return m.Owners
	}
	return nil
}

func (m *Params) GetRoles() []string {
	if m != nil {
		return m.Roles
	}
	return nil
}

func (m *Params) GetWfId() string {
	if m != nil {
		return m.WfId
	}
	return ""
}

func (m *Params) GetDatabase() string {
	if m != nil {
		return m.Database
	}
	return ""
}

func (m *Params) GetFirstMonth() string {
	if m != nil {
		return m.FirstMonth
	}
	return ""
}

// 文件参数
type FileParams struct {
	FilePath             string   `protobuf:"bytes,1,opt,name=file_path,json=filePath,proto3" json:"file_path"`
	ZipFilePath          string   `protobuf:"bytes,2,opt,name=zip_file_path,json=zipFilePath,proto3" json:"zip_file_path"`
	PayFilePath          string   `protobuf:"bytes,3,opt,name=pay_file_path,json=payFilePath,proto3" json:"pay_file_path"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FileParams) Reset()         { *m = FileParams{} }
func (m *FileParams) String() string { return proto.CompactTextString(m) }
func (*FileParams) ProtoMessage()    {}
func (*FileParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_91b94b655bd2a7e5, []int{1}
}

func (m *FileParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FileParams.Unmarshal(m, b)
}
func (m *FileParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FileParams.Marshal(b, m, deterministic)
}
func (m *FileParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileParams.Merge(m, src)
}
func (m *FileParams) XXX_Size() int {
	return xxx_messageInfo_FileParams.Size(m)
}
func (m *FileParams) XXX_DiscardUnknown() {
	xxx_messageInfo_FileParams.DiscardUnknown(m)
}

var xxx_messageInfo_FileParams proto.InternalMessageInfo

func (m *FileParams) GetFilePath() string {
	if m != nil {
		return m.FilePath
	}
	return ""
}

func (m *FileParams) GetZipFilePath() string {
	if m != nil {
		return m.ZipFilePath
	}
	return ""
}

func (m *FileParams) GetPayFilePath() string {
	if m != nil {
		return m.PayFilePath
	}
	return ""
}

// 基础参数
type MappingParams struct {
	JobId                string   `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	MappingId            string   `protobuf:"bytes,2,opt,name=mapping_id,json=mappingId,proto3" json:"mapping_id"`
	UserId               string   `protobuf:"bytes,3,opt,name=user_id,json=userId,proto3" json:"user_id"`
	AppId                string   `protobuf:"bytes,4,opt,name=app_id,json=appId,proto3" json:"app_id"`
	Lang                 string   `protobuf:"bytes,5,opt,name=lang,proto3" json:"lang"`
	Domain               string   `protobuf:"bytes,6,opt,name=domain,proto3" json:"domain"`
	DatastoreId          string   `protobuf:"bytes,7,opt,name=datastore_id,json=datastoreId,proto3" json:"datastore_id"`
	EmptyChange          bool     `protobuf:"varint,12,opt,name=empty_change,json=emptyChange,proto3" json:"empty_change"`
	AccessKeys           []string `protobuf:"bytes,8,rep,name=access_keys,json=accessKeys,proto3" json:"access_keys"`
	Owners               []string `protobuf:"bytes,9,rep,name=owners,proto3" json:"owners"`
	Roles                []string `protobuf:"bytes,10,rep,name=roles,proto3" json:"roles"`
	Database             string   `protobuf:"bytes,11,opt,name=database,proto3" json:"database"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MappingParams) Reset()         { *m = MappingParams{} }
func (m *MappingParams) String() string { return proto.CompactTextString(m) }
func (*MappingParams) ProtoMessage()    {}
func (*MappingParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_91b94b655bd2a7e5, []int{2}
}

func (m *MappingParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MappingParams.Unmarshal(m, b)
}
func (m *MappingParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MappingParams.Marshal(b, m, deterministic)
}
func (m *MappingParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MappingParams.Merge(m, src)
}
func (m *MappingParams) XXX_Size() int {
	return xxx_messageInfo_MappingParams.Size(m)
}
func (m *MappingParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MappingParams.DiscardUnknown(m)
}

var xxx_messageInfo_MappingParams proto.InternalMessageInfo

func (m *MappingParams) GetJobId() string {
	if m != nil {
		return m.JobId
	}
	return ""
}

func (m *MappingParams) GetMappingId() string {
	if m != nil {
		return m.MappingId
	}
	return ""
}

func (m *MappingParams) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *MappingParams) GetAppId() string {
	if m != nil {
		return m.AppId
	}
	return ""
}

func (m *MappingParams) GetLang() string {
	if m != nil {
		return m.Lang
	}
	return ""
}

func (m *MappingParams) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

func (m *MappingParams) GetDatastoreId() string {
	if m != nil {
		return m.DatastoreId
	}
	return ""
}

func (m *MappingParams) GetEmptyChange() bool {
	if m != nil {
		return m.EmptyChange
	}
	return false
}

func (m *MappingParams) GetAccessKeys() []string {
	if m != nil {
		return m.AccessKeys
	}
	return nil
}

func (m *MappingParams) GetOwners() []string {
	if m != nil {
		return m.Owners
	}
	return nil
}

func (m *MappingParams) GetRoles() []string {
	if m != nil {
		return m.Roles
	}
	return nil
}

func (m *MappingParams) GetDatabase() string {
	if m != nil {
		return m.Database
	}
	return ""
}

type MappingRequest struct {
	BaseParams           *MappingParams `protobuf:"bytes,1,opt,name=base_params,json=baseParams,proto3" json:"base_params"`
	FilePath             string         `protobuf:"bytes,2,opt,name=file_path,json=filePath,proto3" json:"file_path"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *MappingRequest) Reset()         { *m = MappingRequest{} }
func (m *MappingRequest) String() string { return proto.CompactTextString(m) }
func (*MappingRequest) ProtoMessage()    {}
func (*MappingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_91b94b655bd2a7e5, []int{3}
}

func (m *MappingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MappingRequest.Unmarshal(m, b)
}
func (m *MappingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MappingRequest.Marshal(b, m, deterministic)
}
func (m *MappingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MappingRequest.Merge(m, src)
}
func (m *MappingRequest) XXX_Size() int {
	return xxx_messageInfo_MappingRequest.Size(m)
}
func (m *MappingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MappingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MappingRequest proto.InternalMessageInfo

func (m *MappingRequest) GetBaseParams() *MappingParams {
	if m != nil {
		return m.BaseParams
	}
	return nil
}

func (m *MappingRequest) GetFilePath() string {
	if m != nil {
		return m.FilePath
	}
	return ""
}

type MappingResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MappingResponse) Reset()         { *m = MappingResponse{} }
func (m *MappingResponse) String() string { return proto.CompactTextString(m) }
func (*MappingResponse) ProtoMessage()    {}
func (*MappingResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_91b94b655bd2a7e5, []int{4}
}

func (m *MappingResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MappingResponse.Unmarshal(m, b)
}
func (m *MappingResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MappingResponse.Marshal(b, m, deterministic)
}
func (m *MappingResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MappingResponse.Merge(m, src)
}
func (m *MappingResponse) XXX_Size() int {
	return xxx_messageInfo_MappingResponse.Size(m)
}
func (m *MappingResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MappingResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MappingResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Params)(nil), "upload.Params")
	proto.RegisterType((*FileParams)(nil), "upload.FileParams")
	proto.RegisterType((*MappingParams)(nil), "upload.MappingParams")
	proto.RegisterType((*MappingRequest)(nil), "upload.MappingRequest")
	proto.RegisterType((*MappingResponse)(nil), "upload.MappingResponse")
}

func init() { proto.RegisterFile("upload.proto", fileDescriptor_91b94b655bd2a7e5) }

var fileDescriptor_91b94b655bd2a7e5 = []byte{
	// 534 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x54, 0xdd, 0x6e, 0x1a, 0x3d,
	0x10, 0xfd, 0xf8, 0x5b, 0x96, 0x59, 0x48, 0xbe, 0xb8, 0x4d, 0xe2, 0xa6, 0x6a, 0x4b, 0xf7, 0x2a,
	0x57, 0xb9, 0x48, 0xa5, 0x3e, 0x40, 0x23, 0x45, 0x5a, 0x55, 0x91, 0x2a, 0xa2, 0x5e, 0x23, 0xb3,
	0x1e, 0xc0, 0x29, 0xd8, 0x8e, 0x6d, 0x8a, 0xc8, 0x5b, 0xb6, 0x4f, 0x54, 0xd9, 0xde, 0x10, 0x48,
	0x28, 0xbd, 0xe3, 0x9c, 0x19, 0x66, 0x8e, 0xe7, 0x1c, 0x80, 0xee, 0x42, 0xcf, 0x14, 0xe3, 0x17,
	0xda, 0x28, 0xa7, 0x48, 0x12, 0x51, 0xfe, 0xab, 0x01, 0xc9, 0x37, 0x66, 0xd8, 0xdc, 0x92, 0x63,
	0x48, 0xee, 0xd4, 0x68, 0x28, 0x38, 0xad, 0xf5, 0x6b, 0xe7, 0x9d, 0x41, 0xeb, 0x4e, 0x8d, 0x0a,
	0x4e, 0x4e, 0x20, 0x61, 0xa5, 0x13, 0x4a, 0xd2, 0x7a, 0xa0, 0x2b, 0x44, 0xce, 0x20, 0x45, 0x59,
	0x2a, 0x2e, 0xe4, 0x84, 0x36, 0x42, 0x65, 0x8d, 0xc9, 0x07, 0xc8, 0x1e, 0x84, 0x1e, 0x96, 0x53,
	0x66, 0x2c, 0x3a, 0xfa, 0x7f, 0x28, 0xc3, 0x83, 0xd0, 0x57, 0x91, 0x21, 0xa7, 0xd0, 0x5e, 0x58,
	0x34, 0x7e, 0x59, 0x33, 0x4e, 0xf5, 0xb0, 0xe0, 0x5e, 0x04, 0xd3, 0xda, 0xf3, 0xad, 0x28, 0x82,
	0x69, 0x5d, 0x70, 0x42, 0xa0, 0x39, 0x63, 0x72, 0x42, 0x93, 0x40, 0x86, 0xcf, 0x5e, 0x18, 0x57,
	0x73, 0x26, 0x24, 0x6d, 0xc7, 0x11, 0x11, 0x91, 0x8f, 0xd0, 0xe5, 0xcc, 0x31, 0xeb, 0x94, 0x41,
	0x3f, 0x28, 0x0d, 0xd5, 0x6c, 0xcd, 0x15, 0x9c, 0xbc, 0x81, 0x74, 0x62, 0xd4, 0x22, 0xec, 0xc9,
	0x42, 0xb9, 0x1d, 0x70, 0xc1, 0xfd, 0xb7, 0x71, 0xae, 0xdd, 0xca, 0x8b, 0x97, 0x13, 0xa4, 0x47,
	0xfd, 0xda, 0x79, 0x3a, 0xc8, 0x02, 0x77, 0x15, 0x28, 0xff, 0x3a, 0x56, 0x96, 0x68, 0xed, 0xf0,
	0x07, 0xae, 0x2c, 0xed, 0xf6, 0x1b, 0xfe, 0x75, 0x91, 0xfa, 0x8a, 0x2b, 0xeb, 0x95, 0xa9, 0xa5,
	0x44, 0x63, 0x69, 0x2f, 0xd4, 0x2a, 0x44, 0x5e, 0x43, 0xcb, 0xa8, 0x19, 0x5a, 0x7a, 0x10, 0xe8,
	0x08, 0xc8, 0x2b, 0x68, 0x2d, 0xc7, 0x5e, 0xc9, 0x61, 0x7c, 0xdc, 0x72, 0x5c, 0x70, 0x7f, 0x5d,
	0x2f, 0x78, 0xc4, 0x2c, 0x52, 0x88, 0xd7, 0x7d, 0xc4, 0xe4, 0x3d, 0xc0, 0x58, 0x18, 0xeb, 0x6e,
	0x94, 0x74, 0x53, 0x4a, 0xe2, 0x71, 0x9f, 0x98, 0xfc, 0x1e, 0xe0, 0x5a, 0xcc, 0xb0, 0xb2, 0xf5,
	0x2d, 0x74, 0xc6, 0x62, 0x86, 0x43, 0xcd, 0xdc, 0xb4, 0x72, 0x36, 0x1d, 0x87, 0xb2, 0x9b, 0x92,
	0x1c, 0x7a, 0xde, 0xa8, 0xa7, 0x86, 0xe8, 0xb1, 0x77, 0xef, 0x7a, 0xa3, 0x47, 0xb3, 0xd5, 0x46,
	0x4f, 0x74, 0x3b, 0xd3, 0x6c, 0xf5, 0xd8, 0x93, 0xff, 0xae, 0x43, 0xef, 0x86, 0x69, 0x2d, 0xe4,
	0x64, 0x7f, 0x9a, 0xde, 0x01, 0xcc, 0x63, 0x9f, 0x2f, 0xc5, 0x6d, 0x9d, 0x8a, 0x29, 0xf8, 0x66,
	0x2e, 0x1a, 0x7f, 0xc9, 0x45, 0x73, 0x57, 0x2e, 0x5a, 0x3b, 0x73, 0x91, 0xec, 0xcd, 0x45, 0xfb,
	0x65, 0x2e, 0x9e, 0x9b, 0xdf, 0xfd, 0xa7, 0xf9, 0xe9, 0x1e, 0xf3, 0x3b, 0xbb, 0xcd, 0x87, 0x4d,
	0xf3, 0x37, 0x7d, 0xce, 0xb6, 0x7d, 0xce, 0x11, 0x0e, 0xaa, 0x9b, 0x0e, 0xf0, 0x7e, 0x81, 0xd6,
	0x91, 0xcf, 0x90, 0xf9, 0xca, 0x50, 0x87, 0x1b, 0x87, 0xcb, 0x66, 0x97, 0xc7, 0x17, 0xd5, 0x2f,
	0x7b, 0xcb, 0x80, 0x01, 0xf8, 0xce, 0x5d, 0x19, 0xa8, 0x6f, 0x67, 0x20, 0x3f, 0x82, 0xc3, 0xf5,
	0x1a, 0xab, 0x95, 0xb4, 0x78, 0x79, 0x0b, 0xbd, 0xef, 0x61, 0xe6, 0x2d, 0x9a, 0x9f, 0xa2, 0x44,
	0xf2, 0x65, 0x6d, 0x6f, 0xe4, 0xc9, 0xc9, 0xb3, 0xa5, 0x95, 0xc2, 0xb3, 0xd3, 0x17, 0x7c, 0x1c,
	0x99, 0xff, 0x37, 0x4a, 0xc2, 0x3f, 0xcf, 0xa7, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x19, 0xa0,
	0xf8, 0x99, 0x89, 0x04, 0x00, 0x00,
}
