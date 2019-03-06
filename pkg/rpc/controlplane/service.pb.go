// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/rpc/controlplane/service.proto

package controlplane

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Graph struct {
	Nodes                []*Node  `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty"`
	Edges                []*Edge  `protobuf:"bytes,2,rep,name=edges,proto3" json:"edges,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Graph) Reset()         { *m = Graph{} }
func (m *Graph) String() string { return proto.CompactTextString(m) }
func (*Graph) ProtoMessage()    {}
func (*Graph) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_b691eb88210b2414, []int{0}
}
func (m *Graph) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Graph.Unmarshal(m, b)
}
func (m *Graph) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Graph.Marshal(b, m, deterministic)
}
func (dst *Graph) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Graph.Merge(dst, src)
}
func (m *Graph) XXX_Size() int {
	return xxx_messageInfo_Graph.Size(m)
}
func (m *Graph) XXX_DiscardUnknown() {
	xxx_messageInfo_Graph.DiscardUnknown(m)
}

var xxx_messageInfo_Graph proto.InternalMessageInfo

func (m *Graph) GetNodes() []*Node {
	if m != nil {
		return m.Nodes
	}
	return nil
}

func (m *Graph) GetEdges() []*Edge {
	if m != nil {
		return m.Edges
	}
	return nil
}

type Node struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Node) Reset()         { *m = Node{} }
func (m *Node) String() string { return proto.CompactTextString(m) }
func (*Node) ProtoMessage()    {}
func (*Node) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_b691eb88210b2414, []int{1}
}
func (m *Node) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Node.Unmarshal(m, b)
}
func (m *Node) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Node.Marshal(b, m, deterministic)
}
func (dst *Node) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Node.Merge(dst, src)
}
func (m *Node) XXX_Size() int {
	return xxx_messageInfo_Node.Size(m)
}
func (m *Node) XXX_DiscardUnknown() {
	xxx_messageInfo_Node.DiscardUnknown(m)
}

var xxx_messageInfo_Node proto.InternalMessageInfo

func (m *Node) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type Edge struct {
	Source               string   `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
	Destination          string   `protobuf:"bytes,2,opt,name=destination,proto3" json:"destination,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Edge) Reset()         { *m = Edge{} }
func (m *Edge) String() string { return proto.CompactTextString(m) }
func (*Edge) ProtoMessage()    {}
func (*Edge) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_b691eb88210b2414, []int{2}
}
func (m *Edge) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Edge.Unmarshal(m, b)
}
func (m *Edge) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Edge.Marshal(b, m, deterministic)
}
func (dst *Edge) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Edge.Merge(dst, src)
}
func (m *Edge) XXX_Size() int {
	return xxx_messageInfo_Edge.Size(m)
}
func (m *Edge) XXX_DiscardUnknown() {
	xxx_messageInfo_Edge.DiscardUnknown(m)
}

var xxx_messageInfo_Edge proto.InternalMessageInfo

func (m *Edge) GetSource() string {
	if m != nil {
		return m.Source
	}
	return ""
}

func (m *Edge) GetDestination() string {
	if m != nil {
		return m.Destination
	}
	return ""
}

type Run struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	CreatedAt            string   `protobuf:"bytes,2,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Run) Reset()         { *m = Run{} }
func (m *Run) String() string { return proto.CompactTextString(m) }
func (*Run) ProtoMessage()    {}
func (*Run) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_b691eb88210b2414, []int{3}
}
func (m *Run) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Run.Unmarshal(m, b)
}
func (m *Run) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Run.Marshal(b, m, deterministic)
}
func (dst *Run) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Run.Merge(dst, src)
}
func (m *Run) XXX_Size() int {
	return xxx_messageInfo_Run.Size(m)
}
func (m *Run) XXX_DiscardUnknown() {
	xxx_messageInfo_Run.DiscardUnknown(m)
}

var xxx_messageInfo_Run proto.InternalMessageInfo

func (m *Run) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Run) GetCreatedAt() string {
	if m != nil {
		return m.CreatedAt
	}
	return ""
}

type ListRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListRequest) Reset()         { *m = ListRequest{} }
func (m *ListRequest) String() string { return proto.CompactTextString(m) }
func (*ListRequest) ProtoMessage()    {}
func (*ListRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_b691eb88210b2414, []int{4}
}
func (m *ListRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListRequest.Unmarshal(m, b)
}
func (m *ListRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListRequest.Marshal(b, m, deterministic)
}
func (dst *ListRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListRequest.Merge(dst, src)
}
func (m *ListRequest) XXX_Size() int {
	return xxx_messageInfo_ListRequest.Size(m)
}
func (m *ListRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListRequest proto.InternalMessageInfo

type ListResponse struct {
	Runs                 []*Run   `protobuf:"bytes,1,rep,name=runs,proto3" json:"runs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListResponse) Reset()         { *m = ListResponse{} }
func (m *ListResponse) String() string { return proto.CompactTextString(m) }
func (*ListResponse) ProtoMessage()    {}
func (*ListResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_b691eb88210b2414, []int{5}
}
func (m *ListResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListResponse.Unmarshal(m, b)
}
func (m *ListResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListResponse.Marshal(b, m, deterministic)
}
func (dst *ListResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListResponse.Merge(dst, src)
}
func (m *ListResponse) XXX_Size() int {
	return xxx_messageInfo_ListResponse.Size(m)
}
func (m *ListResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListResponse proto.InternalMessageInfo

func (m *ListResponse) GetRuns() []*Run {
	if m != nil {
		return m.Runs
	}
	return nil
}

func init() {
	proto.RegisterType((*Graph)(nil), "tempo.adagio.controlplane.Graph")
	proto.RegisterType((*Node)(nil), "tempo.adagio.controlplane.Node")
	proto.RegisterType((*Edge)(nil), "tempo.adagio.controlplane.Edge")
	proto.RegisterType((*Run)(nil), "tempo.adagio.controlplane.Run")
	proto.RegisterType((*ListRequest)(nil), "tempo.adagio.controlplane.ListRequest")
	proto.RegisterType((*ListResponse)(nil), "tempo.adagio.controlplane.ListResponse")
}

func init() {
	proto.RegisterFile("pkg/rpc/controlplane/service.proto", fileDescriptor_service_b691eb88210b2414)
}

var fileDescriptor_service_b691eb88210b2414 = []byte{
	// 317 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0xcf, 0x4b, 0x2b, 0x31,
	0x10, 0xc7, 0xd9, 0xed, 0xb6, 0xd0, 0x69, 0x5f, 0x0f, 0x73, 0x78, 0xec, 0x2b, 0x3c, 0x5d, 0x72,
	0xd0, 0x9e, 0xb6, 0x50, 0xf5, 0xae, 0x15, 0x11, 0x41, 0x44, 0xe2, 0x41, 0xf0, 0x22, 0x71, 0x33,
	0xac, 0xc1, 0x36, 0x89, 0x49, 0xd6, 0x3f, 0xcc, 0xbf, 0x50, 0xf6, 0x47, 0x61, 0x2f, 0xb6, 0xb7,
	0x9d, 0xf9, 0x7e, 0x3e, 0x9b, 0x4c, 0x12, 0x60, 0xf6, 0xa3, 0x5c, 0x3a, 0x5b, 0x2c, 0x0b, 0xa3,
	0x83, 0x33, 0x1b, 0xbb, 0x11, 0x9a, 0x96, 0x9e, 0xdc, 0x97, 0x2a, 0x28, 0xb7, 0xce, 0x04, 0x83,
	0xff, 0x02, 0x6d, 0xad, 0xc9, 0x85, 0x14, 0xa5, 0x32, 0x79, 0x1f, 0x64, 0x15, 0x0c, 0x6f, 0x9d,
	0xb0, 0xef, 0x78, 0x01, 0x43, 0x6d, 0x24, 0xf9, 0x34, 0xca, 0x06, 0x8b, 0xc9, 0xea, 0x38, 0xff,
	0xd5, 0xc9, 0x1f, 0x8c, 0x24, 0xde, 0xd2, 0xb5, 0x46, 0xb2, 0x24, 0x9f, 0xc6, 0x07, 0xb5, 0x1b,
	0x59, 0x12, 0x6f, 0x69, 0x36, 0x87, 0xa4, 0xfe, 0x0b, 0x22, 0x24, 0x5a, 0x6c, 0x29, 0x8d, 0xb2,
	0x68, 0x31, 0xe6, 0xcd, 0x37, 0xbb, 0x84, 0xa4, 0x46, 0xf1, 0x2f, 0x8c, 0xbc, 0xa9, 0x5c, 0xb1,
	0x4b, 0xbb, 0x0a, 0x33, 0x98, 0x48, 0xf2, 0x41, 0x69, 0x11, 0x94, 0xd1, 0x69, 0xdc, 0x84, 0xfd,
	0x16, 0x3b, 0x87, 0x01, 0xaf, 0x34, 0xce, 0x20, 0x56, 0xb2, 0x93, 0x63, 0x25, 0xf1, 0x3f, 0x40,
	0xe1, 0x48, 0x04, 0x92, 0xaf, 0x22, 0x74, 0xde, 0xb8, 0xeb, 0x5c, 0x05, 0xf6, 0x07, 0x26, 0xf7,
	0xca, 0x07, 0x4e, 0x9f, 0x15, 0xf9, 0xc0, 0xd6, 0x30, 0x6d, 0x4b, 0x6f, 0x8d, 0xf6, 0x84, 0x2b,
	0x48, 0x5c, 0xa5, 0x77, 0xe7, 0x73, 0xb4, 0x67, 0x50, 0x5e, 0x69, 0xde, 0xb0, 0xab, 0xef, 0x08,
	0xa6, 0xd7, 0x6d, 0xf4, 0x58, 0x47, 0x78, 0x07, 0xc3, 0xa7, 0x20, 0x5c, 0xc0, 0x6c, 0x8f, 0xdf,
	0x5c, 0xc8, 0xfc, 0xc0, 0x0a, 0xf8, 0x0c, 0x49, 0xbd, 0x3f, 0x3c, 0xd9, 0xc3, 0xf5, 0xe6, 0x99,
	0x9f, 0x1e, 0xe4, 0xda, 0x41, 0xd7, 0xb3, 0x97, 0x69, 0x3f, 0x7c, 0x1b, 0x35, 0x8f, 0xe8, 0xec,
	0x27, 0x00, 0x00, 0xff, 0xff, 0x3f, 0xc1, 0x99, 0x9a, 0x6a, 0x02, 0x00, 0x00,
}
