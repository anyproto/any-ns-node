// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/anyns_api.proto

package anyns_api

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type NameAvailableRequest struct {
	// Name including .any suffix
	FullName string `protobuf:"bytes,1,opt,name=fullName,proto3" json:"fullName,omitempty"`
}

func (m *NameAvailableRequest) Reset()         { *m = NameAvailableRequest{} }
func (m *NameAvailableRequest) String() string { return proto.CompactTextString(m) }
func (*NameAvailableRequest) ProtoMessage()    {}
func (*NameAvailableRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_75814d9f891937a2, []int{0}
}
func (m *NameAvailableRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NameAvailableRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NameAvailableRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NameAvailableRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NameAvailableRequest.Merge(m, src)
}
func (m *NameAvailableRequest) XXX_Size() int {
	return m.Size()
}
func (m *NameAvailableRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_NameAvailableRequest.DiscardUnknown(m)
}

var xxx_messageInfo_NameAvailableRequest proto.InternalMessageInfo

func (m *NameAvailableRequest) GetFullName() string {
	if m != nil {
		return m.FullName
	}
	return ""
}

type NameByAddressRequest struct {
	// An Ethereum address that owns that name
	OwnerEthAddress string `protobuf:"bytes,1,opt,name=ownerEthAddress,proto3" json:"ownerEthAddress,omitempty"`
}

func (m *NameByAddressRequest) Reset()         { *m = NameByAddressRequest{} }
func (m *NameByAddressRequest) String() string { return proto.CompactTextString(m) }
func (*NameByAddressRequest) ProtoMessage()    {}
func (*NameByAddressRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_75814d9f891937a2, []int{1}
}
func (m *NameByAddressRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NameByAddressRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NameByAddressRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NameByAddressRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NameByAddressRequest.Merge(m, src)
}
func (m *NameByAddressRequest) XXX_Size() int {
	return m.Size()
}
func (m *NameByAddressRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_NameByAddressRequest.DiscardUnknown(m)
}

var xxx_messageInfo_NameByAddressRequest proto.InternalMessageInfo

func (m *NameByAddressRequest) GetOwnerEthAddress() string {
	if m != nil {
		return m.OwnerEthAddress
	}
	return ""
}

type NameAvailableResponse struct {
	Available bool `protobuf:"varint,1,opt,name=available,proto3" json:"available,omitempty"`
	// An Ethereum address that owns that name
	// This field is non-empty only if name is "already registered"
	OwnerEthAddress string `protobuf:"bytes,2,opt,name=ownerEthAddress,proto3" json:"ownerEthAddress,omitempty"`
	// A content hash attached to this name
	// This field is non-empty only if name is "already registered"
	OwnerAnyAddress string `protobuf:"bytes,3,opt,name=ownerAnyAddress,proto3" json:"ownerAnyAddress,omitempty"`
	// A SpaceID attached to this name
	// This field is non-empty only if name is "already registered"
	SpaceId string `protobuf:"bytes,4,opt,name=spaceId,proto3" json:"spaceId,omitempty"`
	// doestn't work with marashalling/unmarshalling
	//google.protobuf.Timestamp nameExpires = 5 [(gogoproto.stdtime) = true];
	NameExpires int64 `protobuf:"varint,5,opt,name=nameExpires,proto3" json:"nameExpires,omitempty"`
}

func (m *NameAvailableResponse) Reset()         { *m = NameAvailableResponse{} }
func (m *NameAvailableResponse) String() string { return proto.CompactTextString(m) }
func (*NameAvailableResponse) ProtoMessage()    {}
func (*NameAvailableResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_75814d9f891937a2, []int{2}
}
func (m *NameAvailableResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NameAvailableResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NameAvailableResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NameAvailableResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NameAvailableResponse.Merge(m, src)
}
func (m *NameAvailableResponse) XXX_Size() int {
	return m.Size()
}
func (m *NameAvailableResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_NameAvailableResponse.DiscardUnknown(m)
}

var xxx_messageInfo_NameAvailableResponse proto.InternalMessageInfo

func (m *NameAvailableResponse) GetAvailable() bool {
	if m != nil {
		return m.Available
	}
	return false
}

func (m *NameAvailableResponse) GetOwnerEthAddress() string {
	if m != nil {
		return m.OwnerEthAddress
	}
	return ""
}

func (m *NameAvailableResponse) GetOwnerAnyAddress() string {
	if m != nil {
		return m.OwnerAnyAddress
	}
	return ""
}

func (m *NameAvailableResponse) GetSpaceId() string {
	if m != nil {
		return m.SpaceId
	}
	return ""
}

func (m *NameAvailableResponse) GetNameExpires() int64 {
	if m != nil {
		return m.NameExpires
	}
	return 0
}

type NameByAddressResponse struct {
	Found bool   `protobuf:"varint,1,opt,name=found,proto3" json:"found,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (m *NameByAddressResponse) Reset()         { *m = NameByAddressResponse{} }
func (m *NameByAddressResponse) String() string { return proto.CompactTextString(m) }
func (*NameByAddressResponse) ProtoMessage()    {}
func (*NameByAddressResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_75814d9f891937a2, []int{3}
}
func (m *NameByAddressResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NameByAddressResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NameByAddressResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NameByAddressResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NameByAddressResponse.Merge(m, src)
}
func (m *NameByAddressResponse) XXX_Size() int {
	return m.Size()
}
func (m *NameByAddressResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_NameByAddressResponse.DiscardUnknown(m)
}

var xxx_messageInfo_NameByAddressResponse proto.InternalMessageInfo

func (m *NameByAddressResponse) GetFound() bool {
	if m != nil {
		return m.Found
	}
	return false
}

func (m *NameByAddressResponse) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterType((*NameAvailableRequest)(nil), "NameAvailableRequest")
	proto.RegisterType((*NameByAddressRequest)(nil), "NameByAddressRequest")
	proto.RegisterType((*NameAvailableResponse)(nil), "NameAvailableResponse")
	proto.RegisterType((*NameByAddressResponse)(nil), "NameByAddressResponse")
}

func init() { proto.RegisterFile("proto/anyns_api.proto", fileDescriptor_75814d9f891937a2) }

var fileDescriptor_75814d9f891937a2 = []byte{
	// 324 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2d, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0x4f, 0xcc, 0xab, 0xcc, 0x2b, 0x8e, 0x4f, 0x2c, 0xc8, 0xd4, 0x03, 0xf3, 0x95, 0x8c,
	0xb8, 0x44, 0xfc, 0x12, 0x73, 0x53, 0x1d, 0xcb, 0x12, 0x33, 0x73, 0x12, 0x93, 0x72, 0x52, 0x83,
	0x52, 0x0b, 0x4b, 0x53, 0x8b, 0x4b, 0x84, 0xa4, 0xb8, 0x38, 0xd2, 0x4a, 0x73, 0x72, 0x40, 0x72,
	0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x70, 0xbe, 0x92, 0x03, 0x44, 0x8f, 0x53, 0xa5, 0x63,
	0x4a, 0x4a, 0x51, 0x6a, 0x71, 0x31, 0x4c, 0x8f, 0x06, 0x17, 0x7f, 0x7e, 0x79, 0x5e, 0x6a, 0x91,
	0x6b, 0x49, 0x06, 0x54, 0x06, 0xaa, 0x15, 0x5d, 0x58, 0xe9, 0x28, 0x23, 0x97, 0x28, 0x9a, 0xb5,
	0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9, 0x42, 0x32, 0x5c, 0x9c, 0x89, 0x30, 0x41, 0xb0, 0x6e, 0x8e,
	0x20, 0x84, 0x00, 0x36, 0x1b, 0x98, 0xb0, 0xda, 0x00, 0x57, 0xe9, 0x98, 0x07, 0x73, 0xa5, 0x04,
	0x33, 0x92, 0x4a, 0x84, 0xb0, 0x90, 0x04, 0x17, 0x7b, 0x71, 0x41, 0x62, 0x72, 0xaa, 0x67, 0x8a,
	0x04, 0x0b, 0x58, 0x05, 0x8c, 0x2b, 0xa4, 0xc0, 0xc5, 0x9d, 0x97, 0x98, 0x9b, 0xea, 0x5a, 0x51,
	0x90, 0x59, 0x94, 0x5a, 0x2c, 0xc1, 0xaa, 0xc0, 0xa8, 0xc1, 0x1c, 0x84, 0x2c, 0xa4, 0xe4, 0x08,
	0xf1, 0x06, 0x52, 0x48, 0x40, 0xbd, 0x21, 0xc2, 0xc5, 0x9a, 0x96, 0x5f, 0x9a, 0x97, 0x02, 0xf5,
	0x02, 0x84, 0x23, 0x24, 0xc4, 0xc5, 0x02, 0xd2, 0x0d, 0x75, 0x33, 0x98, 0x6d, 0x34, 0x81, 0x91,
	0x8b, 0xd5, 0x11, 0x14, 0x29, 0x42, 0x4e, 0x5c, 0xfc, 0x9e, 0xc5, 0x28, 0xa1, 0x22, 0x24, 0xaa,
	0x87, 0x2d, 0x72, 0xa4, 0xc4, 0xf4, 0xb0, 0x06, 0x9e, 0x12, 0x83, 0x90, 0x33, 0x97, 0x80, 0x7b,
	0x6a, 0x09, 0x8a, 0x9b, 0xa0, 0x86, 0xa0, 0xc7, 0x16, 0xd4, 0x10, 0x0c, 0xa7, 0x2b, 0x31, 0x38,
	0xa9, 0x9d, 0x78, 0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c, 0x13, 0x1e,
	0xcb, 0x31, 0x5c, 0x78, 0x2c, 0xc7, 0x70, 0xe3, 0xb1, 0x1c, 0x43, 0x14, 0x4f, 0x41, 0x12, 0x22,
	0x05, 0x25, 0xb1, 0x81, 0x93, 0x90, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xc5, 0xf2, 0x7f, 0xb0,
	0x5b, 0x02, 0x00, 0x00,
}

func (m *NameAvailableRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NameAvailableRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NameAvailableRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.FullName) > 0 {
		i -= len(m.FullName)
		copy(dAtA[i:], m.FullName)
		i = encodeVarintAnynsApi(dAtA, i, uint64(len(m.FullName)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *NameByAddressRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NameByAddressRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NameByAddressRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.OwnerEthAddress) > 0 {
		i -= len(m.OwnerEthAddress)
		copy(dAtA[i:], m.OwnerEthAddress)
		i = encodeVarintAnynsApi(dAtA, i, uint64(len(m.OwnerEthAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *NameAvailableResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NameAvailableResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NameAvailableResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.NameExpires != 0 {
		i = encodeVarintAnynsApi(dAtA, i, uint64(m.NameExpires))
		i--
		dAtA[i] = 0x28
	}
	if len(m.SpaceId) > 0 {
		i -= len(m.SpaceId)
		copy(dAtA[i:], m.SpaceId)
		i = encodeVarintAnynsApi(dAtA, i, uint64(len(m.SpaceId)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.OwnerAnyAddress) > 0 {
		i -= len(m.OwnerAnyAddress)
		copy(dAtA[i:], m.OwnerAnyAddress)
		i = encodeVarintAnynsApi(dAtA, i, uint64(len(m.OwnerAnyAddress)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.OwnerEthAddress) > 0 {
		i -= len(m.OwnerEthAddress)
		copy(dAtA[i:], m.OwnerEthAddress)
		i = encodeVarintAnynsApi(dAtA, i, uint64(len(m.OwnerEthAddress)))
		i--
		dAtA[i] = 0x12
	}
	if m.Available {
		i--
		if m.Available {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *NameByAddressResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NameByAddressResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NameByAddressResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintAnynsApi(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0x12
	}
	if m.Found {
		i--
		if m.Found {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintAnynsApi(dAtA []byte, offset int, v uint64) int {
	offset -= sovAnynsApi(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *NameAvailableRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.FullName)
	if l > 0 {
		n += 1 + l + sovAnynsApi(uint64(l))
	}
	return n
}

func (m *NameByAddressRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.OwnerEthAddress)
	if l > 0 {
		n += 1 + l + sovAnynsApi(uint64(l))
	}
	return n
}

func (m *NameAvailableResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Available {
		n += 2
	}
	l = len(m.OwnerEthAddress)
	if l > 0 {
		n += 1 + l + sovAnynsApi(uint64(l))
	}
	l = len(m.OwnerAnyAddress)
	if l > 0 {
		n += 1 + l + sovAnynsApi(uint64(l))
	}
	l = len(m.SpaceId)
	if l > 0 {
		n += 1 + l + sovAnynsApi(uint64(l))
	}
	if m.NameExpires != 0 {
		n += 1 + sovAnynsApi(uint64(m.NameExpires))
	}
	return n
}

func (m *NameByAddressResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Found {
		n += 2
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovAnynsApi(uint64(l))
	}
	return n
}

func sovAnynsApi(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozAnynsApi(x uint64) (n int) {
	return sovAnynsApi(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *NameAvailableRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAnynsApi
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: NameAvailableRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NameAvailableRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FullName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAnynsApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FullName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAnynsApi(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *NameByAddressRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAnynsApi
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: NameByAddressRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NameByAddressRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OwnerEthAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAnynsApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OwnerEthAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAnynsApi(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *NameAvailableResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAnynsApi
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: NameAvailableResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NameAvailableResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Available", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Available = bool(v != 0)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OwnerEthAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAnynsApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OwnerEthAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OwnerAnyAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAnynsApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OwnerAnyAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpaceId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAnynsApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SpaceId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NameExpires", wireType)
			}
			m.NameExpires = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NameExpires |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipAnynsApi(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *NameByAddressResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAnynsApi
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: NameByAddressResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NameByAddressResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Found", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Found = bool(v != 0)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAnynsApi
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAnynsApi(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAnynsApi
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipAnynsApi(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAnynsApi
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAnynsApi
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthAnynsApi
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupAnynsApi
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthAnynsApi
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthAnynsApi        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAnynsApi          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupAnynsApi = fmt.Errorf("proto: unexpected end of group")
)
