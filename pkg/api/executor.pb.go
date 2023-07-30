// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.22.5
// source: executor.proto

package api

import (
	context "context"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
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

type RunJobRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OnFireLogID  uint32            `protobuf:"varint,1,opt,name=onFireLogID,proto3" json:"onFireLogID,omitempty"`
	Job          *Job              `protobuf:"bytes,2,opt,name=job,proto3" json:"job,omitempty"`
	TraceContext map[string]string `protobuf:"bytes,3,rep,name=traceContext,proto3" json:"traceContext,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *RunJobRequest) Reset() {
	*x = RunJobRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_executor_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunJobRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunJobRequest) ProtoMessage() {}

func (x *RunJobRequest) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunJobRequest.ProtoReflect.Descriptor instead.
func (*RunJobRequest) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{0}
}

func (x *RunJobRequest) GetOnFireLogID() uint32 {
	if x != nil {
		return x.OnFireLogID
	}
	return 0
}

func (x *RunJobRequest) GetJob() *Job {
	if x != nil {
		return x.Job
	}
	return nil
}

func (x *RunJobRequest) GetTraceContext() map[string]string {
	if x != nil {
		return x.TraceContext
	}
	return nil
}

type Job struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GlueType                 string            `protobuf:"bytes,1,opt,name=glueType,proto3" json:"glueType,omitempty"`
	Source                   map[string]string `protobuf:"bytes,2,rep,name=source,proto3" json:"source,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Param                    map[string]string `protobuf:"bytes,3,rep,name=param,proto3" json:"param,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ExecutorExecuteTimeoutMs int64             `protobuf:"varint,4,opt,name=executorExecuteTimeoutMs,proto3" json:"executorExecuteTimeoutMs,omitempty"`
}

func (x *Job) Reset() {
	*x = Job{}
	if protoimpl.UnsafeEnabled {
		mi := &file_executor_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Job) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Job) ProtoMessage() {}

func (x *Job) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Job.ProtoReflect.Descriptor instead.
func (*Job) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{1}
}

func (x *Job) GetGlueType() string {
	if x != nil {
		return x.GlueType
	}
	return ""
}

func (x *Job) GetSource() map[string]string {
	if x != nil {
		return x.Source
	}
	return nil
}

func (x *Job) GetParam() map[string]string {
	if x != nil {
		return x.Param
	}
	return nil
}

func (x *Job) GetExecutorExecuteTimeoutMs() int64 {
	if x != nil {
		return x.ExecutorExecuteTimeoutMs
	}
	return 0
}

type RunJobResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OnFireLogID  uint32            `protobuf:"varint,1,opt,name=onFireLogID,proto3" json:"onFireLogID,omitempty"`
	Result       *JobResult        `protobuf:"bytes,2,opt,name=result,proto3" json:"result,omitempty"`
	TraceContext map[string]string `protobuf:"bytes,3,rep,name=traceContext,proto3" json:"traceContext,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *RunJobResponse) Reset() {
	*x = RunJobResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_executor_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunJobResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunJobResponse) ProtoMessage() {}

func (x *RunJobResponse) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunJobResponse.ProtoReflect.Descriptor instead.
func (*RunJobResponse) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{2}
}

func (x *RunJobResponse) GetOnFireLogID() uint32 {
	if x != nil {
		return x.OnFireLogID
	}
	return 0
}

func (x *RunJobResponse) GetResult() *JobResult {
	if x != nil {
		return x.Result
	}
	return nil
}

func (x *RunJobResponse) GetTraceContext() map[string]string {
	if x != nil {
		return x.TraceContext
	}
	return nil
}

type JobResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok     bool   `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
	Err    string `protobuf:"bytes,2,opt,name=err,proto3" json:"err,omitempty"`
	Result string `protobuf:"bytes,3,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *JobResult) Reset() {
	*x = JobResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_executor_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JobResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JobResult) ProtoMessage() {}

func (x *JobResult) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JobResult.ProtoReflect.Descriptor instead.
func (*JobResult) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{3}
}

func (x *JobResult) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

func (x *JobResult) GetErr() string {
	if x != nil {
		return x.Err
	}
	return ""
}

func (x *JobResult) GetResult() string {
	if x != nil {
		return x.Result
	}
	return ""
}

type HeartBeatRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HeartBeatRequest) Reset() {
	*x = HeartBeatRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_executor_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeartBeatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartBeatRequest) ProtoMessage() {}

func (x *HeartBeatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartBeatRequest.ProtoReflect.Descriptor instead.
func (*HeartBeatRequest) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{4}
}

type HeartBeatResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HealthStatus *HealthStatus `protobuf:"bytes,1,opt,name=healthStatus,proto3" json:"healthStatus,omitempty"`
}

func (x *HeartBeatResponse) Reset() {
	*x = HeartBeatResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_executor_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeartBeatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartBeatResponse) ProtoMessage() {}

func (x *HeartBeatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeartBeatResponse.ProtoReflect.Descriptor instead.
func (*HeartBeatResponse) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{5}
}

func (x *HeartBeatResponse) GetHealthStatus() *HealthStatus {
	if x != nil {
		return x.HealthStatus
	}
	return nil
}

type HealthStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Workload        float32 `protobuf:"fixed32,1,opt,name=workload,proto3" json:"workload,omitempty"` //当前负载情况
	GracefulStopped bool    `protobuf:"varint,2,opt,name=gracefulStopped,proto3" json:"gracefulStopped,omitempty"`
}

func (x *HealthStatus) Reset() {
	*x = HealthStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_executor_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HealthStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthStatus) ProtoMessage() {}

func (x *HealthStatus) ProtoReflect() protoreflect.Message {
	mi := &file_executor_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthStatus.ProtoReflect.Descriptor instead.
func (*HealthStatus) Descriptor() ([]byte, []int) {
	return file_executor_proto_rawDescGZIP(), []int{6}
}

func (x *HealthStatus) GetWorkload() float32 {
	if x != nil {
		return x.Workload
	}
	return 0
}

func (x *HealthStatus) GetGracefulStopped() bool {
	if x != nil {
		return x.GracefulStopped
	}
	return false
}

var File_executor_proto protoreflect.FileDescriptor

var file_executor_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x03, 0x61, 0x70, 0x69, 0x22, 0xd8, 0x01, 0x0a, 0x0d, 0x52, 0x75, 0x6e, 0x4a, 0x6f, 0x62,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x6f, 0x6e, 0x46, 0x69, 0x72,
	0x65, 0x4c, 0x6f, 0x67, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x6f, 0x6e,
	0x46, 0x69, 0x72, 0x65, 0x4c, 0x6f, 0x67, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x03, 0x6a, 0x6f, 0x62,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4a, 0x6f, 0x62,
	0x52, 0x03, 0x6a, 0x6f, 0x62, 0x12, 0x48, 0x0a, 0x0c, 0x74, 0x72, 0x61, 0x63, 0x65, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x78, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x52, 0x75, 0x6e, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e,
	0x54, 0x72, 0x61, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x0c, 0x74, 0x72, 0x61, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x1a,
	0x3f, 0x0a, 0x11, 0x54, 0x72, 0x61, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x22, 0xab, 0x02, 0x0a, 0x03, 0x4a, 0x6f, 0x62, 0x12, 0x1a, 0x0a, 0x08, 0x67, 0x6c, 0x75, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x67, 0x6c, 0x75, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4a, 0x6f, 0x62, 0x2e, 0x53,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x12, 0x29, 0x0a, 0x05, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x13, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4a, 0x6f, 0x62, 0x2e, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x12, 0x3a, 0x0a,
	0x18, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x6f, 0x72, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65,
	0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x4d, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x18, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x6f, 0x72, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65,
	0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x4d, 0x73, 0x1a, 0x39, 0x0a, 0x0b, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x1a, 0x38, 0x0a, 0x0a, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xe6,
	0x01, 0x0a, 0x0e, 0x52, 0x75, 0x6e, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x20, 0x0a, 0x0b, 0x6f, 0x6e, 0x46, 0x69, 0x72, 0x65, 0x4c, 0x6f, 0x67, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x6f, 0x6e, 0x46, 0x69, 0x72, 0x65, 0x4c, 0x6f,
	0x67, 0x49, 0x44, 0x12, 0x26, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x49, 0x0a, 0x0c, 0x74,
	0x72, 0x61, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x25, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52, 0x75, 0x6e, 0x4a, 0x6f, 0x62, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x54, 0x72, 0x61, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x78, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0c, 0x74, 0x72, 0x61, 0x63, 0x65, 0x43,
	0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x1a, 0x3f, 0x0a, 0x11, 0x54, 0x72, 0x61, 0x63, 0x65, 0x43,
	0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x45, 0x0a, 0x09, 0x4a, 0x6f, 0x62, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x02, 0x6f, 0x6b, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x72, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x65, 0x72, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x12,
	0x0a, 0x10, 0x48, 0x65, 0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x22, 0x4a, 0x0a, 0x11, 0x48, 0x65, 0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x35, 0x0a, 0x0c, 0x68, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x0c, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x54,
	0x0a, 0x0c, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1a,
	0x0a, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x02,
	0x52, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x28, 0x0a, 0x0f, 0x67, 0x72,
	0x61, 0x63, 0x65, 0x66, 0x75, 0x6c, 0x53, 0x74, 0x6f, 0x70, 0x70, 0x65, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x0f, 0x67, 0x72, 0x61, 0x63, 0x65, 0x66, 0x75, 0x6c, 0x53, 0x74, 0x6f,
	0x70, 0x70, 0x65, 0x64, 0x32, 0x7d, 0x0a, 0x08, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x6f, 0x72,
	0x12, 0x3a, 0x0a, 0x09, 0x48, 0x65, 0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x12, 0x15, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x48, 0x65, 0x61, 0x72, 0x74, 0x42, 0x65, 0x61, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x48, 0x65, 0x61, 0x72, 0x74,
	0x42, 0x65, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x35, 0x0a, 0x06,
	0x52, 0x75, 0x6e, 0x4a, 0x6f, 0x62, 0x12, 0x12, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52, 0x75, 0x6e,
	0x4a, 0x6f, 0x62, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x52, 0x75, 0x6e, 0x4a, 0x6f, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28,
	0x01, 0x30, 0x01, 0x42, 0x20, 0x5a, 0x1e, 0x73, 0x75, 0x70, 0x65, 0x72, 0x6e, 0x6f, 0x76, 0x61,
	0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x6b, 0x69, 0x74, 0x65, 0x78, 0x5f, 0x67, 0x65,
	0x6e, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_executor_proto_rawDescOnce sync.Once
	file_executor_proto_rawDescData = file_executor_proto_rawDesc
)

func file_executor_proto_rawDescGZIP() []byte {
	file_executor_proto_rawDescOnce.Do(func() {
		file_executor_proto_rawDescData = protoimpl.X.CompressGZIP(file_executor_proto_rawDescData)
	})
	return file_executor_proto_rawDescData
}

var file_executor_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_executor_proto_goTypes = []interface{}{
	(*RunJobRequest)(nil),     // 0: api.RunJobRequest
	(*Job)(nil),               // 1: api.Job
	(*RunJobResponse)(nil),    // 2: api.RunJobResponse
	(*JobResult)(nil),         // 3: api.JobResult
	(*HeartBeatRequest)(nil),  // 4: api.HeartBeatRequest
	(*HeartBeatResponse)(nil), // 5: api.HeartBeatResponse
	(*HealthStatus)(nil),      // 6: api.HealthStatus
	nil,                       // 7: api.RunJobRequest.TraceContextEntry
	nil,                       // 8: api.Job.SourceEntry
	nil,                       // 9: api.Job.ParamEntry
	nil,                       // 10: api.RunJobResponse.TraceContextEntry
}
var file_executor_proto_depIdxs = []int32{
	1,  // 0: api.RunJobRequest.job:type_name -> api.Job
	7,  // 1: api.RunJobRequest.traceContext:type_name -> api.RunJobRequest.TraceContextEntry
	8,  // 2: api.Job.source:type_name -> api.Job.SourceEntry
	9,  // 3: api.Job.param:type_name -> api.Job.ParamEntry
	3,  // 4: api.RunJobResponse.result:type_name -> api.JobResult
	10, // 5: api.RunJobResponse.traceContext:type_name -> api.RunJobResponse.TraceContextEntry
	6,  // 6: api.HeartBeatResponse.healthStatus:type_name -> api.HealthStatus
	4,  // 7: api.Executor.HeartBeat:input_type -> api.HeartBeatRequest
	0,  // 8: api.Executor.RunJob:input_type -> api.RunJobRequest
	5,  // 9: api.Executor.HeartBeat:output_type -> api.HeartBeatResponse
	2,  // 10: api.Executor.RunJob:output_type -> api.RunJobResponse
	9,  // [9:11] is the sub-list for method output_type
	7,  // [7:9] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_executor_proto_init() }
func file_executor_proto_init() {
	if File_executor_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_executor_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunJobRequest); i {
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
		file_executor_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Job); i {
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
		file_executor_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunJobResponse); i {
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
		file_executor_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JobResult); i {
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
		file_executor_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeartBeatRequest); i {
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
		file_executor_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeartBeatResponse); i {
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
		file_executor_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HealthStatus); i {
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
			RawDescriptor: file_executor_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_executor_proto_goTypes,
		DependencyIndexes: file_executor_proto_depIdxs,
		MessageInfos:      file_executor_proto_msgTypes,
	}.Build()
	File_executor_proto = out.File
	file_executor_proto_rawDesc = nil
	file_executor_proto_goTypes = nil
	file_executor_proto_depIdxs = nil
}

var _ context.Context

// Code generated by Kitex v0.6.1. DO NOT EDIT.

type Executor interface {
	HeartBeat(ctx context.Context, req *HeartBeatRequest) (res *HeartBeatResponse, err error)
	RunJob(stream Executor_RunJobServer) (err error)
}

type Executor_RunJobServer interface {
	streaming.Stream
	Recv() (*RunJobRequest, error)
	Send(*RunJobResponse) error
}
