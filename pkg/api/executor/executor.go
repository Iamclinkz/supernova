// Code generated by Kitex v0.6.1. DO NOT EDIT.

package executor

import (
	"context"
	"fmt"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
	proto "google.golang.org/protobuf/proto"
	"supernova/pkg/api"
)

func serviceInfo() *kitex.ServiceInfo {
	return executorServiceInfo
}

var executorServiceInfo = NewServiceInfo()

func NewServiceInfo() *kitex.ServiceInfo {
	serviceName := "Executor"
	handlerType := (*api.Executor)(nil)
	methods := map[string]kitex.MethodInfo{
		"HeartBeat": kitex.NewMethodInfo(heartBeatHandler, newHeartBeatArgs, newHeartBeatResult, false),
		"RunJob":    kitex.NewMethodInfo(runJobHandler, newRunJobArgs, newRunJobResult, false),
	}
	extra := map[string]interface{}{
		"PackageName": "api",
	}
	extra["streaming"] = true
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         methods,
		PayloadCodec:    kitex.Protobuf,
		KiteXGenVersion: "v0.6.1",
		Extra:           extra,
	}
	return svcInfo
}

func heartBeatHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new(api.HeartBeatRequest)
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.(api.Executor).HeartBeat(ctx, req)
		if err != nil {
			return err
		}
		if err := st.SendMsg(resp); err != nil {
			return err
		}
	case *HeartBeatArgs:
		success, err := handler.(api.Executor).HeartBeat(ctx, s.Req)
		if err != nil {
			return err
		}
		realResult := result.(*HeartBeatResult)
		realResult.Success = success
	}
	return nil
}
func newHeartBeatArgs() interface{} {
	return &HeartBeatArgs{}
}

func newHeartBeatResult() interface{} {
	return &HeartBeatResult{}
}

type HeartBeatArgs struct {
	Req *api.HeartBeatRequest
}

func (p *HeartBeatArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(api.HeartBeatRequest)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *HeartBeatArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *HeartBeatArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *HeartBeatArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, fmt.Errorf("No req in HeartBeatArgs")
	}
	return proto.Marshal(p.Req)
}

func (p *HeartBeatArgs) Unmarshal(in []byte) error {
	msg := new(api.HeartBeatRequest)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var HeartBeatArgs_Req_DEFAULT *api.HeartBeatRequest

func (p *HeartBeatArgs) GetReq() *api.HeartBeatRequest {
	if !p.IsSetReq() {
		return HeartBeatArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *HeartBeatArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *HeartBeatArgs) GetFirstArgument() interface{} {
	return p.Req
}

type HeartBeatResult struct {
	Success *api.HeartBeatResponse
}

var HeartBeatResult_Success_DEFAULT *api.HeartBeatResponse

func (p *HeartBeatResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(api.HeartBeatResponse)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *HeartBeatResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *HeartBeatResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *HeartBeatResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, fmt.Errorf("No req in HeartBeatResult")
	}
	return proto.Marshal(p.Success)
}

func (p *HeartBeatResult) Unmarshal(in []byte) error {
	msg := new(api.HeartBeatResponse)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *HeartBeatResult) GetSuccess() *api.HeartBeatResponse {
	if !p.IsSetSuccess() {
		return HeartBeatResult_Success_DEFAULT
	}
	return p.Success
}

func (p *HeartBeatResult) SetSuccess(x interface{}) {
	p.Success = x.(*api.HeartBeatResponse)
}

func (p *HeartBeatResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *HeartBeatResult) GetResult() interface{} {
	return p.Success
}

func runJobHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st := arg.(*streaming.Args).Stream
	stream := &executorRunJobServer{st}
	return handler.(api.Executor).RunJob(stream)
}

type executorRunJobClient struct {
	streaming.Stream
}

func (x *executorRunJobClient) Send(m *api.RunJobRequest) error {
	return x.Stream.SendMsg(m)
}
func (x *executorRunJobClient) Recv() (*api.RunJobResponse, error) {
	m := new(api.RunJobResponse)
	return m, x.Stream.RecvMsg(m)
}

type executorRunJobServer struct {
	streaming.Stream
}

func (x *executorRunJobServer) Send(m *api.RunJobResponse) error {
	return x.Stream.SendMsg(m)
}

func (x *executorRunJobServer) Recv() (*api.RunJobRequest, error) {
	m := new(api.RunJobRequest)
	return m, x.Stream.RecvMsg(m)
}

func newRunJobArgs() interface{} {
	return &RunJobArgs{}
}

func newRunJobResult() interface{} {
	return &RunJobResult{}
}

type RunJobArgs struct {
	Req *api.RunJobRequest
}

func (p *RunJobArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(api.RunJobRequest)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *RunJobArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *RunJobArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *RunJobArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, fmt.Errorf("No req in RunJobArgs")
	}
	return proto.Marshal(p.Req)
}

func (p *RunJobArgs) Unmarshal(in []byte) error {
	msg := new(api.RunJobRequest)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var RunJobArgs_Req_DEFAULT *api.RunJobRequest

func (p *RunJobArgs) GetReq() *api.RunJobRequest {
	if !p.IsSetReq() {
		return RunJobArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *RunJobArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *RunJobArgs) GetFirstArgument() interface{} {
	return p.Req
}

type RunJobResult struct {
	Success *api.RunJobResponse
}

var RunJobResult_Success_DEFAULT *api.RunJobResponse

func (p *RunJobResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(api.RunJobResponse)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *RunJobResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *RunJobResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *RunJobResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, fmt.Errorf("No req in RunJobResult")
	}
	return proto.Marshal(p.Success)
}

func (p *RunJobResult) Unmarshal(in []byte) error {
	msg := new(api.RunJobResponse)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *RunJobResult) GetSuccess() *api.RunJobResponse {
	if !p.IsSetSuccess() {
		return RunJobResult_Success_DEFAULT
	}
	return p.Success
}

func (p *RunJobResult) SetSuccess(x interface{}) {
	p.Success = x.(*api.RunJobResponse)
}

func (p *RunJobResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *RunJobResult) GetResult() interface{} {
	return p.Success
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) HeartBeat(ctx context.Context, Req *api.HeartBeatRequest) (r *api.HeartBeatResponse, err error) {
	var _args HeartBeatArgs
	_args.Req = Req
	var _result HeartBeatResult
	if err = p.c.Call(ctx, "HeartBeat", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) RunJob(ctx context.Context) (Executor_RunJobClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "RunJob", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &executorRunJobClient{res.Stream}
	return stream, nil
}
