// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: everlast/vauth/v1/vauth.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

// ProofExternalOwnedAccount store the proof that account is external owned account (EOA)
type ProofExternalOwnedAccount struct {
	// account is the cosmos bech32 address of the account that has proof
	Account string `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	// hash is the keccak256 of the message that was signed on
	Hash string `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	// signature is the signed message using private key
	Signature string `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *ProofExternalOwnedAccount) Reset()         { *m = ProofExternalOwnedAccount{} }
func (m *ProofExternalOwnedAccount) String() string { return proto.CompactTextString(m) }
func (*ProofExternalOwnedAccount) ProtoMessage()    {}
func (*ProofExternalOwnedAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_0e506c933a2ca858, []int{0}
}
func (m *ProofExternalOwnedAccount) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ProofExternalOwnedAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ProofExternalOwnedAccount.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ProofExternalOwnedAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProofExternalOwnedAccount.Merge(m, src)
}
func (m *ProofExternalOwnedAccount) XXX_Size() int {
	return m.Size()
}
func (m *ProofExternalOwnedAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_ProofExternalOwnedAccount.DiscardUnknown(m)
}

var xxx_messageInfo_ProofExternalOwnedAccount proto.InternalMessageInfo

func (m *ProofExternalOwnedAccount) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *ProofExternalOwnedAccount) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *ProofExternalOwnedAccount) GetSignature() string {
	if m != nil {
		return m.Signature
	}
	return ""
}

func init() {
	proto.RegisterType((*ProofExternalOwnedAccount)(nil), "everlast.vauth.v1.ProofExternalOwnedAccount")
}

func init() { proto.RegisterFile("everlast/vauth/v1/vauth.proto", fileDescriptor_0e506c933a2ca858) }

var fileDescriptor_0e506c933a2ca858 = []byte{
	// 213 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4d, 0x2d, 0x4b, 0x2d,
	0xca, 0x49, 0x2c, 0x2e, 0xd1, 0x2f, 0x4b, 0x2c, 0x2d, 0xc9, 0xd0, 0x2f, 0x33, 0x84, 0x30, 0xf4,
	0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0x04, 0x61, 0xd2, 0x7a, 0x10, 0xd1, 0x32, 0x43, 0x29, 0x91,
	0xf4, 0xfc, 0xf4, 0x7c, 0xb0, 0xac, 0x3e, 0x88, 0x05, 0x51, 0xa8, 0x94, 0xce, 0x25, 0x19, 0x50,
	0x94, 0x9f, 0x9f, 0xe6, 0x5a, 0x51, 0x92, 0x5a, 0x94, 0x97, 0x98, 0xe3, 0x5f, 0x9e, 0x97, 0x9a,
	0xe2, 0x98, 0x9c, 0x9c, 0x5f, 0x9a, 0x57, 0x22, 0x24, 0xc1, 0xc5, 0x9e, 0x08, 0x61, 0x4a, 0x30,
	0x2a, 0x30, 0x6a, 0x70, 0x06, 0xc1, 0xb8, 0x42, 0x42, 0x5c, 0x2c, 0x19, 0x89, 0xc5, 0x19, 0x12,
	0x4c, 0x60, 0x61, 0x30, 0x5b, 0x48, 0x86, 0x8b, 0xb3, 0x38, 0x33, 0x3d, 0x2f, 0xb1, 0xa4, 0xb4,
	0x28, 0x55, 0x82, 0x19, 0x2c, 0x81, 0x10, 0x70, 0x72, 0x3e, 0xf1, 0x48, 0x8e, 0xf1, 0xc2, 0x23,
	0x39, 0xc6, 0x07, 0x8f, 0xe4, 0x18, 0x27, 0x3c, 0x96, 0x63, 0xb8, 0xf0, 0x58, 0x8e, 0xe1, 0xc6,
	0x63, 0x39, 0x86, 0x28, 0xcd, 0xf4, 0xcc, 0x92, 0x8c, 0xd2, 0x24, 0xbd, 0xe4, 0xfc, 0x5c, 0x7d,
	0xd7, 0xe2, 0xe4, 0xc4, 0x3c, 0x27, 0x57, 0x7d, 0xb8, 0xef, 0x2a, 0xa0, 0xfe, 0x2b, 0xa9, 0x2c,
	0x48, 0x2d, 0x4e, 0x62, 0x03, 0x3b, 0xda, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0xfc, 0xb5, 0x52,
	0x0d, 0xfe, 0x00, 0x00, 0x00,
}

func (m *ProofExternalOwnedAccount) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProofExternalOwnedAccount) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ProofExternalOwnedAccount) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Signature) > 0 {
		i -= len(m.Signature)
		copy(dAtA[i:], m.Signature)
		i = encodeVarintVauth(dAtA, i, uint64(len(m.Signature)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Hash) > 0 {
		i -= len(m.Hash)
		copy(dAtA[i:], m.Hash)
		i = encodeVarintVauth(dAtA, i, uint64(len(m.Hash)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Account) > 0 {
		i -= len(m.Account)
		copy(dAtA[i:], m.Account)
		i = encodeVarintVauth(dAtA, i, uint64(len(m.Account)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintVauth(dAtA []byte, offset int, v uint64) int {
	offset -= sovVauth(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ProofExternalOwnedAccount) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Account)
	if l > 0 {
		n += 1 + l + sovVauth(uint64(l))
	}
	l = len(m.Hash)
	if l > 0 {
		n += 1 + l + sovVauth(uint64(l))
	}
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovVauth(uint64(l))
	}
	return n
}

func sovVauth(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozVauth(x uint64) (n int) {
	return sovVauth(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ProofExternalOwnedAccount) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowVauth
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
			return fmt.Errorf("proto: ProofExternalOwnedAccount: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProofExternalOwnedAccount: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Account", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVauth
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
				return ErrInvalidLengthVauth
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVauth
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Account = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hash", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVauth
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
				return ErrInvalidLengthVauth
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVauth
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Hash = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVauth
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
				return ErrInvalidLengthVauth
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVauth
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipVauth(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthVauth
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
func skipVauth(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowVauth
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
					return 0, ErrIntOverflowVauth
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
					return 0, ErrIntOverflowVauth
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
				return 0, ErrInvalidLengthVauth
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupVauth
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthVauth
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthVauth        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowVauth          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupVauth = fmt.Errorf("proto: unexpected end of group")
)
