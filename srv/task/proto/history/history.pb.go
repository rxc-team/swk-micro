// Code generated by protoc-gen-go. DO NOT EDIT.
// source: history.proto

package history

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

// 任务数据
type Download struct {
	JobId                string   `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	JobName              string   `protobuf:"bytes,2,opt,name=job_name,json=jobName,proto3" json:"job_name"`
	Origin               string   `protobuf:"bytes,3,opt,name=origin,proto3" json:"origin"`
	UserId               string   `protobuf:"bytes,4,opt,name=user_id,json=userId,proto3" json:"user_id"`
	Progress             int64    `protobuf:"varint,5,opt,name=progress,proto3" json:"progress"`
	StartTime            string   `protobuf:"bytes,6,opt,name=start_time,json=startTime,proto3" json:"start_time"`
	EndTime              string   `protobuf:"bytes,11,opt,name=end_time,json=endTime,proto3" json:"end_time"`
	Message              string   `protobuf:"bytes,7,opt,name=message,proto3" json:"message"`
	ErrorFilePath        string   `protobuf:"bytes,8,opt,name=error_file_path,json=errorFilePath,proto3" json:"error_file_path"`
	FilePath             string   `protobuf:"bytes,12,opt,name=file_path,json=filePath,proto3" json:"file_path"`
	CurrentStep          string   `protobuf:"bytes,9,opt,name=current_step,json=currentStep,proto3" json:"current_step"`
	ScheduleId           string   `protobuf:"bytes,10,opt,name=schedule_id,json=scheduleId,proto3" json:"schedule_id"`
	Steps                []string `protobuf:"bytes,14,rep,name=steps,proto3" json:"steps"`
	TaskType             string   `protobuf:"bytes,13,opt,name=task_type,json=taskType,proto3" json:"task_type"`
	AppId                string   `protobuf:"bytes,15,opt,name=app_id,json=appId,proto3" json:"app_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Download) Reset()         { *m = Download{} }
func (m *Download) String() string { return proto.CompactTextString(m) }
func (*Download) ProtoMessage()    {}
func (*Download) Descriptor() ([]byte, []int) {
	return fileDescriptor_454388b49b309873, []int{0}
}

func (m *Download) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Download.Unmarshal(m, b)
}
func (m *Download) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Download.Marshal(b, m, deterministic)
}
func (m *Download) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Download.Merge(m, src)
}
func (m *Download) XXX_Size() int {
	return xxx_messageInfo_Download.Size(m)
}
func (m *Download) XXX_DiscardUnknown() {
	xxx_messageInfo_Download.DiscardUnknown(m)
}

var xxx_messageInfo_Download proto.InternalMessageInfo

func (m *Download) GetJobId() string {
	if m != nil {
		return m.JobId
	}
	return ""
}

func (m *Download) GetJobName() string {
	if m != nil {
		return m.JobName
	}
	return ""
}

func (m *Download) GetOrigin() string {
	if m != nil {
		return m.Origin
	}
	return ""
}

func (m *Download) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *Download) GetProgress() int64 {
	if m != nil {
		return m.Progress
	}
	return 0
}

func (m *Download) GetStartTime() string {
	if m != nil {
		return m.StartTime
	}
	return ""
}

func (m *Download) GetEndTime() string {
	if m != nil {
		return m.EndTime
	}
	return ""
}

func (m *Download) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *Download) GetErrorFilePath() string {
	if m != nil {
		return m.ErrorFilePath
	}
	return ""
}

func (m *Download) GetFilePath() string {
	if m != nil {
		return m.FilePath
	}
	return ""
}

func (m *Download) GetCurrentStep() string {
	if m != nil {
		return m.CurrentStep
	}
	return ""
}

func (m *Download) GetScheduleId() string {
	if m != nil {
		return m.ScheduleId
	}
	return ""
}

func (m *Download) GetSteps() []string {
	if m != nil {
		return m.Steps
	}
	return nil
}

func (m *Download) GetTaskType() string {
	if m != nil {
		return m.TaskType
	}
	return ""
}

func (m *Download) GetAppId() string {
	if m != nil {
		return m.AppId
	}
	return ""
}

// 查找多条记录
type DownloadRequest struct {
	UserId               string   `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id"`
	JobId                string   `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	PageIndex            int64    `protobuf:"varint,3,opt,name=page_index,json=pageIndex,proto3" json:"page_index"`
	PageSize             int64    `protobuf:"varint,4,opt,name=page_size,json=pageSize,proto3" json:"page_size"`
	Database             string   `protobuf:"bytes,5,opt,name=database,proto3" json:"database"`
	ScheduleId           string   `protobuf:"bytes,6,opt,name=schedule_id,json=scheduleId,proto3" json:"schedule_id"`
	AppId                string   `protobuf:"bytes,7,opt,name=app_id,json=appId,proto3" json:"app_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DownloadRequest) Reset()         { *m = DownloadRequest{} }
func (m *DownloadRequest) String() string { return proto.CompactTextString(m) }
func (*DownloadRequest) ProtoMessage()    {}
func (*DownloadRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_454388b49b309873, []int{1}
}

func (m *DownloadRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DownloadRequest.Unmarshal(m, b)
}
func (m *DownloadRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DownloadRequest.Marshal(b, m, deterministic)
}
func (m *DownloadRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DownloadRequest.Merge(m, src)
}
func (m *DownloadRequest) XXX_Size() int {
	return xxx_messageInfo_DownloadRequest.Size(m)
}
func (m *DownloadRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DownloadRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DownloadRequest proto.InternalMessageInfo

func (m *DownloadRequest) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *DownloadRequest) GetJobId() string {
	if m != nil {
		return m.JobId
	}
	return ""
}

func (m *DownloadRequest) GetPageIndex() int64 {
	if m != nil {
		return m.PageIndex
	}
	return 0
}

func (m *DownloadRequest) GetPageSize() int64 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

func (m *DownloadRequest) GetDatabase() string {
	if m != nil {
		return m.Database
	}
	return ""
}

func (m *DownloadRequest) GetScheduleId() string {
	if m != nil {
		return m.ScheduleId
	}
	return ""
}

func (m *DownloadRequest) GetAppId() string {
	if m != nil {
		return m.AppId
	}
	return ""
}

type DownloadResponse struct {
	Histories            []*Download `protobuf:"bytes,1,rep,name=histories,proto3" json:"histories"`
	Total                int64       `protobuf:"varint,2,opt,name=total,proto3" json:"total"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *DownloadResponse) Reset()         { *m = DownloadResponse{} }
func (m *DownloadResponse) String() string { return proto.CompactTextString(m) }
func (*DownloadResponse) ProtoMessage()    {}
func (*DownloadResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_454388b49b309873, []int{2}
}

func (m *DownloadResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DownloadResponse.Unmarshal(m, b)
}
func (m *DownloadResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DownloadResponse.Marshal(b, m, deterministic)
}
func (m *DownloadResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DownloadResponse.Merge(m, src)
}
func (m *DownloadResponse) XXX_Size() int {
	return xxx_messageInfo_DownloadResponse.Size(m)
}
func (m *DownloadResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DownloadResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DownloadResponse proto.InternalMessageInfo

func (m *DownloadResponse) GetHistories() []*Download {
	if m != nil {
		return m.Histories
	}
	return nil
}

func (m *DownloadResponse) GetTotal() int64 {
	if m != nil {
		return m.Total
	}
	return 0
}

type Message struct {
	StartTime            string   `protobuf:"bytes,1,opt,name=start_time,json=startTime,proto3" json:"start_time"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_454388b49b309873, []int{3}
}

func (m *Message) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Message.Unmarshal(m, b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Message.Marshal(b, m, deterministic)
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return xxx_messageInfo_Message.Size(m)
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

func (m *Message) GetStartTime() string {
	if m != nil {
		return m.StartTime
	}
	return ""
}

func (m *Message) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

// 任务数据
type History struct {
	JobId                string     `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	JobName              string     `protobuf:"bytes,2,opt,name=job_name,json=jobName,proto3" json:"job_name"`
	Origin               string     `protobuf:"bytes,3,opt,name=origin,proto3" json:"origin"`
	UserId               string     `protobuf:"bytes,4,opt,name=user_id,json=userId,proto3" json:"user_id"`
	Progress             int64      `protobuf:"varint,5,opt,name=progress,proto3" json:"progress"`
	StartTime            string     `protobuf:"bytes,6,opt,name=start_time,json=startTime,proto3" json:"start_time"`
	EndTime              string     `protobuf:"bytes,11,opt,name=end_time,json=endTime,proto3" json:"end_time"`
	Message              []*Message `protobuf:"bytes,7,rep,name=message,proto3" json:"message"`
	ErrorFilePath        string     `protobuf:"bytes,8,opt,name=error_file_path,json=errorFilePath,proto3" json:"error_file_path"`
	FilePath             string     `protobuf:"bytes,12,opt,name=file_path,json=filePath,proto3" json:"file_path"`
	CurrentStep          string     `protobuf:"bytes,9,opt,name=current_step,json=currentStep,proto3" json:"current_step"`
	Steps                []string   `protobuf:"bytes,14,rep,name=steps,proto3" json:"steps"`
	ScheduleId           string     `protobuf:"bytes,10,opt,name=schedule_id,json=scheduleId,proto3" json:"schedule_id"`
	TaskType             string     `protobuf:"bytes,13,opt,name=task_type,json=taskType,proto3" json:"task_type"`
	AppId                string     `protobuf:"bytes,15,opt,name=app_id,json=appId,proto3" json:"app_id"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *History) Reset()         { *m = History{} }
func (m *History) String() string { return proto.CompactTextString(m) }
func (*History) ProtoMessage()    {}
func (*History) Descriptor() ([]byte, []int) {
	return fileDescriptor_454388b49b309873, []int{4}
}

func (m *History) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_History.Unmarshal(m, b)
}
func (m *History) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_History.Marshal(b, m, deterministic)
}
func (m *History) XXX_Merge(src proto.Message) {
	xxx_messageInfo_History.Merge(m, src)
}
func (m *History) XXX_Size() int {
	return xxx_messageInfo_History.Size(m)
}
func (m *History) XXX_DiscardUnknown() {
	xxx_messageInfo_History.DiscardUnknown(m)
}

var xxx_messageInfo_History proto.InternalMessageInfo

func (m *History) GetJobId() string {
	if m != nil {
		return m.JobId
	}
	return ""
}

func (m *History) GetJobName() string {
	if m != nil {
		return m.JobName
	}
	return ""
}

func (m *History) GetOrigin() string {
	if m != nil {
		return m.Origin
	}
	return ""
}

func (m *History) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *History) GetProgress() int64 {
	if m != nil {
		return m.Progress
	}
	return 0
}

func (m *History) GetStartTime() string {
	if m != nil {
		return m.StartTime
	}
	return ""
}

func (m *History) GetEndTime() string {
	if m != nil {
		return m.EndTime
	}
	return ""
}

func (m *History) GetMessage() []*Message {
	if m != nil {
		return m.Message
	}
	return nil
}

func (m *History) GetErrorFilePath() string {
	if m != nil {
		return m.ErrorFilePath
	}
	return ""
}

func (m *History) GetFilePath() string {
	if m != nil {
		return m.FilePath
	}
	return ""
}

func (m *History) GetCurrentStep() string {
	if m != nil {
		return m.CurrentStep
	}
	return ""
}

func (m *History) GetSteps() []string {
	if m != nil {
		return m.Steps
	}
	return nil
}

func (m *History) GetScheduleId() string {
	if m != nil {
		return m.ScheduleId
	}
	return ""
}

func (m *History) GetTaskType() string {
	if m != nil {
		return m.TaskType
	}
	return ""
}

func (m *History) GetAppId() string {
	if m != nil {
		return m.AppId
	}
	return ""
}

// 查找多条记录
type HistoriesRequest struct {
	UserId               string   `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id"`
	JobId                string   `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	PageIndex            int64    `protobuf:"varint,3,opt,name=page_index,json=pageIndex,proto3" json:"page_index"`
	PageSize             int64    `protobuf:"varint,4,opt,name=page_size,json=pageSize,proto3" json:"page_size"`
	Database             string   `protobuf:"bytes,5,opt,name=database,proto3" json:"database"`
	ScheduleId           string   `protobuf:"bytes,6,opt,name=schedule_id,json=scheduleId,proto3" json:"schedule_id"`
	AppId                string   `protobuf:"bytes,7,opt,name=app_id,json=appId,proto3" json:"app_id"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HistoriesRequest) Reset()         { *m = HistoriesRequest{} }
func (m *HistoriesRequest) String() string { return proto.CompactTextString(m) }
func (*HistoriesRequest) ProtoMessage()    {}
func (*HistoriesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_454388b49b309873, []int{5}
}

func (m *HistoriesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HistoriesRequest.Unmarshal(m, b)
}
func (m *HistoriesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HistoriesRequest.Marshal(b, m, deterministic)
}
func (m *HistoriesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HistoriesRequest.Merge(m, src)
}
func (m *HistoriesRequest) XXX_Size() int {
	return xxx_messageInfo_HistoriesRequest.Size(m)
}
func (m *HistoriesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_HistoriesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_HistoriesRequest proto.InternalMessageInfo

func (m *HistoriesRequest) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *HistoriesRequest) GetJobId() string {
	if m != nil {
		return m.JobId
	}
	return ""
}

func (m *HistoriesRequest) GetPageIndex() int64 {
	if m != nil {
		return m.PageIndex
	}
	return 0
}

func (m *HistoriesRequest) GetPageSize() int64 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

func (m *HistoriesRequest) GetDatabase() string {
	if m != nil {
		return m.Database
	}
	return ""
}

func (m *HistoriesRequest) GetScheduleId() string {
	if m != nil {
		return m.ScheduleId
	}
	return ""
}

func (m *HistoriesRequest) GetAppId() string {
	if m != nil {
		return m.AppId
	}
	return ""
}

type HistoriesResponse struct {
	Histories            []*History `protobuf:"bytes,1,rep,name=histories,proto3" json:"histories"`
	Total                int64      `protobuf:"varint,2,opt,name=total,proto3" json:"total"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *HistoriesResponse) Reset()         { *m = HistoriesResponse{} }
func (m *HistoriesResponse) String() string { return proto.CompactTextString(m) }
func (*HistoriesResponse) ProtoMessage()    {}
func (*HistoriesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_454388b49b309873, []int{6}
}

func (m *HistoriesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HistoriesResponse.Unmarshal(m, b)
}
func (m *HistoriesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HistoriesResponse.Marshal(b, m, deterministic)
}
func (m *HistoriesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HistoriesResponse.Merge(m, src)
}
func (m *HistoriesResponse) XXX_Size() int {
	return xxx_messageInfo_HistoriesResponse.Size(m)
}
func (m *HistoriesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_HistoriesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_HistoriesResponse proto.InternalMessageInfo

func (m *HistoriesResponse) GetHistories() []*History {
	if m != nil {
		return m.Histories
	}
	return nil
}

func (m *HistoriesResponse) GetTotal() int64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func init() {
	proto.RegisterType((*Download)(nil), "history.Download")
	proto.RegisterType((*DownloadRequest)(nil), "history.DownloadRequest")
	proto.RegisterType((*DownloadResponse)(nil), "history.DownloadResponse")
	proto.RegisterType((*Message)(nil), "history.Message")
	proto.RegisterType((*History)(nil), "history.History")
	proto.RegisterType((*HistoriesRequest)(nil), "history.HistoriesRequest")
	proto.RegisterType((*HistoriesResponse)(nil), "history.HistoriesResponse")
}

func init() { proto.RegisterFile("history.proto", fileDescriptor_454388b49b309873) }

var fileDescriptor_454388b49b309873 = []byte{
	// 593 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xdc, 0x55, 0xdd, 0x6e, 0xd3, 0x4c,
	0x10, 0xfd, 0x5c, 0x7f, 0x8d, 0xe3, 0x49, 0xd3, 0x9f, 0x15, 0x3f, 0xdb, 0xa2, 0x8a, 0x90, 0x0b,
	0x54, 0x71, 0x51, 0xa4, 0xf2, 0x06, 0x08, 0x55, 0x0d, 0x12, 0x08, 0x39, 0xbd, 0xe1, 0xca, 0xda,
	0x64, 0xa7, 0xc9, 0x86, 0xc4, 0xbb, 0xec, 0x6e, 0x80, 0xf4, 0x81, 0xb8, 0xe6, 0x15, 0x78, 0x04,
	0x78, 0x22, 0xb4, 0x6b, 0x3b, 0x76, 0x7e, 0xa4, 0x4a, 0x5c, 0x20, 0xc1, 0xe5, 0xcc, 0x19, 0xcf,
	0xce, 0x9c, 0x33, 0x47, 0x86, 0xf6, 0x58, 0x18, 0x2b, 0xf5, 0xe2, 0x5c, 0x69, 0x69, 0x25, 0x89,
	0x8a, 0xb0, 0xfb, 0x2d, 0x84, 0xe6, 0x2b, 0xf9, 0x39, 0x9b, 0x4a, 0xc6, 0xc9, 0x7d, 0x68, 0x4c,
	0xe4, 0x20, 0x15, 0x9c, 0x06, 0x9d, 0xe0, 0x2c, 0x4e, 0x76, 0x27, 0x72, 0xd0, 0xe3, 0xe4, 0x18,
	0x9a, 0x2e, 0x9d, 0xb1, 0x19, 0xd2, 0x1d, 0x0f, 0x44, 0x13, 0x39, 0x78, 0xcb, 0x66, 0x48, 0x1e,
	0x40, 0x43, 0x6a, 0x31, 0x12, 0x19, 0x0d, 0x3d, 0x50, 0x44, 0xe4, 0x21, 0x44, 0x73, 0x83, 0xda,
	0xb5, 0xfa, 0x3f, 0x07, 0x5c, 0xd8, 0xe3, 0xe4, 0x04, 0x9a, 0x4a, 0xcb, 0x91, 0x46, 0x63, 0xe8,
	0x6e, 0x27, 0x38, 0x0b, 0x93, 0x65, 0x4c, 0x4e, 0x01, 0x8c, 0x65, 0xda, 0xa6, 0x56, 0xcc, 0x90,
	0x36, 0xfc, 0x77, 0xb1, 0xcf, 0x5c, 0x8b, 0x19, 0xba, 0x31, 0x30, 0xe3, 0x39, 0xd8, 0xca, 0xc7,
	0xc0, 0x8c, 0x7b, 0x88, 0x42, 0x34, 0x43, 0x63, 0xd8, 0x08, 0x69, 0x94, 0x23, 0x45, 0x48, 0x9e,
	0xc2, 0x01, 0x6a, 0x2d, 0x75, 0x7a, 0x23, 0xa6, 0x98, 0x2a, 0x66, 0xc7, 0xb4, 0xe9, 0x2b, 0xda,
	0x3e, 0x7d, 0x29, 0xa6, 0xf8, 0x8e, 0xd9, 0x31, 0x79, 0x04, 0x71, 0x55, 0xb1, 0xe7, 0x2b, 0x9a,
	0x37, 0x25, 0xf8, 0x04, 0xf6, 0x86, 0x73, 0xad, 0x31, 0xb3, 0xa9, 0xb1, 0xa8, 0x68, 0xec, 0xf1,
	0x56, 0x91, 0xeb, 0x5b, 0x54, 0xe4, 0x31, 0xb4, 0xcc, 0x70, 0x8c, 0x7c, 0x3e, 0x45, 0xb7, 0x34,
	0xf8, 0x0a, 0x28, 0x53, 0x3d, 0x4e, 0xee, 0xc1, 0xae, 0xfb, 0xd6, 0xd0, 0xfd, 0x4e, 0xe8, 0xa8,
	0xf5, 0x81, 0x7b, 0xd6, 0x32, 0xf3, 0x21, 0xb5, 0x0b, 0x85, 0xb4, 0x9d, 0x3f, 0xeb, 0x12, 0xd7,
	0x0b, 0x85, 0x4e, 0x0e, 0xa6, 0x94, 0x6b, 0x77, 0x90, 0xcb, 0xc1, 0x94, 0xea, 0xf1, 0xee, 0x8f,
	0x00, 0x0e, 0x4a, 0xc9, 0x12, 0xfc, 0x38, 0x47, 0x63, 0xeb, 0x7c, 0x07, 0x2b, 0x7c, 0x57, 0x92,
	0xee, 0xd4, 0x25, 0x3d, 0x05, 0x50, 0x6c, 0x84, 0xa9, 0xc8, 0x38, 0x7e, 0xf1, 0xda, 0x85, 0x49,
	0xec, 0x32, 0x3d, 0x97, 0x70, 0x63, 0x79, 0xd8, 0x88, 0x5b, 0xf4, 0x02, 0x3a, 0x99, 0xd8, 0x08,
	0xfb, 0xe2, 0x16, 0x9d, 0x84, 0x9c, 0x59, 0x36, 0x60, 0x06, 0xbd, 0x84, 0x71, 0xb2, 0x8c, 0xd7,
	0x69, 0x68, 0x6c, 0xd0, 0x50, 0xed, 0x14, 0xd5, 0x77, 0x7a, 0x0f, 0x87, 0xd5, 0x4a, 0x46, 0xc9,
	0xcc, 0x20, 0x79, 0x0e, 0x71, 0x7e, 0xa5, 0x02, 0x0d, 0x0d, 0x3a, 0xe1, 0x59, 0xeb, 0xe2, 0xe8,
	0xbc, 0x3c, 0xe3, 0x65, 0x75, 0x55, 0xe3, 0x28, 0xb6, 0xd2, 0xb2, 0xa9, 0x5f, 0x35, 0x4c, 0xf2,
	0xa0, 0xfb, 0x12, 0xa2, 0x37, 0xc5, 0x31, 0xac, 0x1e, 0x58, 0xb0, 0x7e, 0x60, 0xb5, 0x2b, 0xda,
	0x59, 0xb9, 0xa2, 0xee, 0xf7, 0x10, 0xa2, 0xab, 0xfc, 0xe5, 0xbf, 0xd7, 0x24, 0xcf, 0xea, 0x26,
	0x71, 0x6c, 0x1e, 0x2e, 0xd9, 0x2c, 0x08, 0xfa, 0xf3, 0xb6, 0xd9, 0xee, 0x8a, 0x3b, 0xcd, 0xf4,
	0x3b, 0xb6, 0xf9, 0x19, 0xc0, 0xe1, 0x55, 0x79, 0x2b, 0xff, 0x8e, 0x6f, 0x8e, 0x6a, 0x3b, 0x15,
	0xc6, 0x39, 0xdf, 0x34, 0x4e, 0x25, 0x75, 0x71, 0xc6, 0x77, 0xfa, 0xe6, 0xe2, 0x6b, 0x00, 0xfb,
	0x45, 0x71, 0x1f, 0xf5, 0x27, 0x31, 0x44, 0x72, 0x05, 0xed, 0x4b, 0x91, 0xf1, 0xe5, 0x8b, 0xe4,
	0x78, 0xad, 0x6d, 0xc5, 0xec, 0xc9, 0xc9, 0x36, 0x28, 0x1f, 0xb0, 0xfb, 0x1f, 0x79, 0x0d, 0x47,
	0xa5, 0x83, 0xab, 0x6e, 0x74, 0xd3, 0xdd, 0x45, 0xb3, 0xe3, 0x2d, 0x48, 0xd9, 0x6b, 0xd0, 0xf0,
	0xbf, 0xb4, 0x17, 0xbf, 0x02, 0x00, 0x00, 0xff, 0xff, 0x4d, 0xed, 0x9b, 0xf1, 0xe3, 0x06, 0x00,
	0x00,
}
