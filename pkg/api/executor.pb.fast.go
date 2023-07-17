// Code generated by Fastpb v0.0.2. DO NOT EDIT.

package api

import (
	fmt "fmt"
	fastpb "github.com/cloudwego/fastpb"
)

var (
	_ = fmt.Errorf
	_ = fastpb.Skip
)

func (x *HeartBeatRequest) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
}

func (x *HeartBeatResponse) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 3:
		offset, err = x.fastReadField3(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_HeartBeatResponse[number], err)
}

func (x *HeartBeatResponse) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.Cpu, offset, err = fastpb.ReadInt32(buf, _type)
	return offset, err
}

func (x *HeartBeatResponse) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	x.Memory, offset, err = fastpb.ReadInt32(buf, _type)
	return offset, err
}

func (x *HeartBeatResponse) fastReadField3(buf []byte, _type int8) (offset int, err error) {
	x.RunningJob, offset, err = fastpb.ReadInt32(buf, _type)
	return offset, err
}

func (x *RunJobRequest) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 3:
		offset, err = x.fastReadField3(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_RunJobRequest[number], err)
}

func (x *RunJobRequest) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.GlueType, offset, err = fastpb.ReadInt32(buf, _type)
	return offset, err
}

func (x *RunJobRequest) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	x.GlueSource, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *RunJobRequest) fastReadField3(buf []byte, _type int8) (offset int, err error) {
	x.ExecutorExecuteTimeoutMs, offset, err = fastpb.ReadInt64(buf, _type)
	return offset, err
}

func (x *RunJobResponse) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 3:
		offset, err = x.fastReadField3(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_RunJobResponse[number], err)
}

func (x *RunJobResponse) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.Ok, offset, err = fastpb.ReadBool(buf, _type)
	return offset, err
}

func (x *RunJobResponse) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	x.Err, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *RunJobResponse) fastReadField3(buf []byte, _type int8) (offset int, err error) {
	x.Result, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *HeartBeatRequest) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	return offset
}

func (x *HeartBeatResponse) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField3(buf[offset:])
	return offset
}

func (x *HeartBeatResponse) fastWriteField1(buf []byte) (offset int) {
	if x.Cpu == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 1, x.GetCpu())
	return offset
}

func (x *HeartBeatResponse) fastWriteField2(buf []byte) (offset int) {
	if x.Memory == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 2, x.GetMemory())
	return offset
}

func (x *HeartBeatResponse) fastWriteField3(buf []byte) (offset int) {
	if x.RunningJob == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 3, x.GetRunningJob())
	return offset
}

func (x *RunJobRequest) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField3(buf[offset:])
	return offset
}

func (x *RunJobRequest) fastWriteField1(buf []byte) (offset int) {
	if x.GlueType == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 1, x.GetGlueType())
	return offset
}

func (x *RunJobRequest) fastWriteField2(buf []byte) (offset int) {
	if x.GlueSource == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 2, x.GetGlueSource())
	return offset
}

func (x *RunJobRequest) fastWriteField3(buf []byte) (offset int) {
	if x.ExecutorExecuteTimeoutMs == 0 {
		return offset
	}
	offset += fastpb.WriteInt64(buf[offset:], 3, x.GetExecutorExecuteTimeoutMs())
	return offset
}

func (x *RunJobResponse) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField3(buf[offset:])
	return offset
}

func (x *RunJobResponse) fastWriteField1(buf []byte) (offset int) {
	if !x.Ok {
		return offset
	}
	offset += fastpb.WriteBool(buf[offset:], 1, x.GetOk())
	return offset
}

func (x *RunJobResponse) fastWriteField2(buf []byte) (offset int) {
	if x.Err == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 2, x.GetErr())
	return offset
}

func (x *RunJobResponse) fastWriteField3(buf []byte) (offset int) {
	if x.Result == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 3, x.GetResult())
	return offset
}

func (x *HeartBeatRequest) Size() (n int) {
	if x == nil {
		return n
	}
	return n
}

func (x *HeartBeatResponse) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	n += x.sizeField3()
	return n
}

func (x *HeartBeatResponse) sizeField1() (n int) {
	if x.Cpu == 0 {
		return n
	}
	n += fastpb.SizeInt32(1, x.GetCpu())
	return n
}

func (x *HeartBeatResponse) sizeField2() (n int) {
	if x.Memory == 0 {
		return n
	}
	n += fastpb.SizeInt32(2, x.GetMemory())
	return n
}

func (x *HeartBeatResponse) sizeField3() (n int) {
	if x.RunningJob == 0 {
		return n
	}
	n += fastpb.SizeInt32(3, x.GetRunningJob())
	return n
}

func (x *RunJobRequest) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	n += x.sizeField3()
	return n
}

func (x *RunJobRequest) sizeField1() (n int) {
	if x.GlueType == 0 {
		return n
	}
	n += fastpb.SizeInt32(1, x.GetGlueType())
	return n
}

func (x *RunJobRequest) sizeField2() (n int) {
	if x.GlueSource == "" {
		return n
	}
	n += fastpb.SizeString(2, x.GetGlueSource())
	return n
}

func (x *RunJobRequest) sizeField3() (n int) {
	if x.ExecutorExecuteTimeoutMs == 0 {
		return n
	}
	n += fastpb.SizeInt64(3, x.GetExecutorExecuteTimeoutMs())
	return n
}

func (x *RunJobResponse) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	n += x.sizeField3()
	return n
}

func (x *RunJobResponse) sizeField1() (n int) {
	if !x.Ok {
		return n
	}
	n += fastpb.SizeBool(1, x.GetOk())
	return n
}

func (x *RunJobResponse) sizeField2() (n int) {
	if x.Err == "" {
		return n
	}
	n += fastpb.SizeString(2, x.GetErr())
	return n
}

func (x *RunJobResponse) sizeField3() (n int) {
	if x.Result == "" {
		return n
	}
	n += fastpb.SizeString(3, x.GetResult())
	return n
}

var fieldIDToName_HeartBeatRequest = map[int32]string{}

var fieldIDToName_HeartBeatResponse = map[int32]string{
	1: "Cpu",
	2: "Memory",
	3: "RunningJob",
}

var fieldIDToName_RunJobRequest = map[int32]string{
	1: "GlueType",
	2: "GlueSource",
	3: "ExecutorExecuteTimeoutMs",
}

var fieldIDToName_RunJobResponse = map[int32]string{
	1: "Ok",
	2: "Err",
	3: "Result",
}
