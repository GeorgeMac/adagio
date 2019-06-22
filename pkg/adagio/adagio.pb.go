// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/adagio/adagio.proto

package adagio

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

type Conclusion int32

const (
	Conclusion_NONE    Conclusion = 0
	Conclusion_SUCCESS Conclusion = 1
	Conclusion_FAIL    Conclusion = 2
	Conclusion_ERROR   Conclusion = 3
)

var Conclusion_name = map[int32]string{
	0: "NONE",
	1: "SUCCESS",
	2: "FAIL",
	3: "ERROR",
}

var Conclusion_value = map[string]int32{
	"NONE":    0,
	"SUCCESS": 1,
	"FAIL":    2,
	"ERROR":   3,
}

func (x Conclusion) String() string {
	return proto.EnumName(Conclusion_name, int32(x))
}

func (Conclusion) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{0}
}

type Event_Type int32

const (
	Event_STATE_TRANSITION Event_Type = 0
)

var Event_Type_name = map[int32]string{
	0: "STATE_TRANSITION",
}

var Event_Type_value = map[string]int32{
	"STATE_TRANSITION": 0,
}

func (x Event_Type) String() string {
	return proto.EnumName(Event_Type_name, int32(x))
}

func (Event_Type) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{1, 0}
}

type Node_Status int32

const (
	Node_WAITING   Node_Status = 0
	Node_READY     Node_Status = 1
	Node_RUNNING   Node_Status = 2
	Node_COMPLETED Node_Status = 3
)

var Node_Status_name = map[int32]string{
	0: "WAITING",
	1: "READY",
	2: "RUNNING",
	3: "COMPLETED",
}

var Node_Status_value = map[string]int32{
	"WAITING":   0,
	"READY":     1,
	"RUNNING":   2,
	"COMPLETED": 3,
}

func (x Node_Status) String() string {
	return proto.EnumName(Node_Status_name, int32(x))
}

func (Node_Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{4, 0}
}

type Run struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	CreatedAt            string   `protobuf:"bytes,2,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Nodes                []*Node  `protobuf:"bytes,3,rep,name=nodes,proto3" json:"nodes,omitempty"`
	Edges                []*Edge  `protobuf:"bytes,4,rep,name=edges,proto3" json:"edges,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Run) Reset()         { *m = Run{} }
func (m *Run) String() string { return proto.CompactTextString(m) }
func (*Run) ProtoMessage()    {}
func (*Run) Descriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{0}
}

func (m *Run) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Run.Unmarshal(m, b)
}
func (m *Run) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Run.Marshal(b, m, deterministic)
}
func (m *Run) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Run.Merge(m, src)
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

func (m *Run) GetNodes() []*Node {
	if m != nil {
		return m.Nodes
	}
	return nil
}

func (m *Run) GetEdges() []*Edge {
	if m != nil {
		return m.Edges
	}
	return nil
}

type Event struct {
	Type                 Event_Type `protobuf:"varint,1,opt,name=type,proto3,enum=adagio.Event_Type" json:"type,omitempty"`
	RunID                string     `protobuf:"bytes,2,opt,name=runID,proto3" json:"runID,omitempty"`
	NodeSpec             *Node_Spec `protobuf:"bytes,3,opt,name=nodeSpec,proto3" json:"nodeSpec,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Event) Reset()         { *m = Event{} }
func (m *Event) String() string { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()    {}
func (*Event) Descriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{1}
}

func (m *Event) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Event.Unmarshal(m, b)
}
func (m *Event) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Event.Marshal(b, m, deterministic)
}
func (m *Event) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Event.Merge(m, src)
}
func (m *Event) XXX_Size() int {
	return xxx_messageInfo_Event.Size(m)
}
func (m *Event) XXX_DiscardUnknown() {
	xxx_messageInfo_Event.DiscardUnknown(m)
}

var xxx_messageInfo_Event proto.InternalMessageInfo

func (m *Event) GetType() Event_Type {
	if m != nil {
		return m.Type
	}
	return Event_STATE_TRANSITION
}

func (m *Event) GetRunID() string {
	if m != nil {
		return m.RunID
	}
	return ""
}

func (m *Event) GetNodeSpec() *Node_Spec {
	if m != nil {
		return m.NodeSpec
	}
	return nil
}

type GraphSpec struct {
	Nodes                []*Node_Spec `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty"`
	Edges                []*Edge      `protobuf:"bytes,2,rep,name=edges,proto3" json:"edges,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *GraphSpec) Reset()         { *m = GraphSpec{} }
func (m *GraphSpec) String() string { return proto.CompactTextString(m) }
func (*GraphSpec) ProtoMessage()    {}
func (*GraphSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{2}
}

func (m *GraphSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GraphSpec.Unmarshal(m, b)
}
func (m *GraphSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GraphSpec.Marshal(b, m, deterministic)
}
func (m *GraphSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GraphSpec.Merge(m, src)
}
func (m *GraphSpec) XXX_Size() int {
	return xxx_messageInfo_GraphSpec.Size(m)
}
func (m *GraphSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_GraphSpec.DiscardUnknown(m)
}

var xxx_messageInfo_GraphSpec proto.InternalMessageInfo

func (m *GraphSpec) GetNodes() []*Node_Spec {
	if m != nil {
		return m.Nodes
	}
	return nil
}

func (m *GraphSpec) GetEdges() []*Edge {
	if m != nil {
		return m.Edges
	}
	return nil
}

type MetadataValue struct {
	Values               []string `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MetadataValue) Reset()         { *m = MetadataValue{} }
func (m *MetadataValue) String() string { return proto.CompactTextString(m) }
func (*MetadataValue) ProtoMessage()    {}
func (*MetadataValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{3}
}

func (m *MetadataValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MetadataValue.Unmarshal(m, b)
}
func (m *MetadataValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MetadataValue.Marshal(b, m, deterministic)
}
func (m *MetadataValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MetadataValue.Merge(m, src)
}
func (m *MetadataValue) XXX_Size() int {
	return xxx_messageInfo_MetadataValue.Size(m)
}
func (m *MetadataValue) XXX_DiscardUnknown() {
	xxx_messageInfo_MetadataValue.DiscardUnknown(m)
}

var xxx_messageInfo_MetadataValue proto.InternalMessageInfo

func (m *MetadataValue) GetValues() []string {
	if m != nil {
		return m.Values
	}
	return nil
}

type Node struct {
	Spec                 *Node_Spec        `protobuf:"bytes,1,opt,name=spec,proto3" json:"spec,omitempty"`
	Status               Node_Status       `protobuf:"varint,2,opt,name=status,proto3,enum=adagio.Node_Status" json:"status,omitempty"`
	Conclusion           Conclusion        `protobuf:"varint,3,opt,name=conclusion,proto3,enum=adagio.Conclusion" json:"conclusion,omitempty"`
	StartedAt            string            `protobuf:"bytes,4,opt,name=started_at,json=startedAt,proto3" json:"started_at,omitempty"`
	FinishedAt           string            `protobuf:"bytes,5,opt,name=finished_at,json=finishedAt,proto3" json:"finished_at,omitempty"`
	Inputs               map[string][]byte `protobuf:"bytes,6,rep,name=inputs,proto3" json:"inputs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Node) Reset()         { *m = Node{} }
func (m *Node) String() string { return proto.CompactTextString(m) }
func (*Node) ProtoMessage()    {}
func (*Node) Descriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{4}
}

func (m *Node) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Node.Unmarshal(m, b)
}
func (m *Node) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Node.Marshal(b, m, deterministic)
}
func (m *Node) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Node.Merge(m, src)
}
func (m *Node) XXX_Size() int {
	return xxx_messageInfo_Node.Size(m)
}
func (m *Node) XXX_DiscardUnknown() {
	xxx_messageInfo_Node.DiscardUnknown(m)
}

var xxx_messageInfo_Node proto.InternalMessageInfo

func (m *Node) GetSpec() *Node_Spec {
	if m != nil {
		return m.Spec
	}
	return nil
}

func (m *Node) GetStatus() Node_Status {
	if m != nil {
		return m.Status
	}
	return Node_WAITING
}

func (m *Node) GetConclusion() Conclusion {
	if m != nil {
		return m.Conclusion
	}
	return Conclusion_NONE
}

func (m *Node) GetStartedAt() string {
	if m != nil {
		return m.StartedAt
	}
	return ""
}

func (m *Node) GetFinishedAt() string {
	if m != nil {
		return m.FinishedAt
	}
	return ""
}

func (m *Node) GetInputs() map[string][]byte {
	if m != nil {
		return m.Inputs
	}
	return nil
}

type Node_Spec struct {
	Name                 string                    `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Runtime              string                    `protobuf:"bytes,2,opt,name=runtime,proto3" json:"runtime,omitempty"`
	Metadata             map[string]*MetadataValue `protobuf:"bytes,3,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *Node_Spec) Reset()         { *m = Node_Spec{} }
func (m *Node_Spec) String() string { return proto.CompactTextString(m) }
func (*Node_Spec) ProtoMessage()    {}
func (*Node_Spec) Descriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{4, 0}
}

func (m *Node_Spec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Node_Spec.Unmarshal(m, b)
}
func (m *Node_Spec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Node_Spec.Marshal(b, m, deterministic)
}
func (m *Node_Spec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Node_Spec.Merge(m, src)
}
func (m *Node_Spec) XXX_Size() int {
	return xxx_messageInfo_Node_Spec.Size(m)
}
func (m *Node_Spec) XXX_DiscardUnknown() {
	xxx_messageInfo_Node_Spec.DiscardUnknown(m)
}

var xxx_messageInfo_Node_Spec proto.InternalMessageInfo

func (m *Node_Spec) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Node_Spec) GetRuntime() string {
	if m != nil {
		return m.Runtime
	}
	return ""
}

func (m *Node_Spec) GetMetadata() map[string]*MetadataValue {
	if m != nil {
		return m.Metadata
	}
	return nil
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
	return fileDescriptor_5eb97351c0f66fbe, []int{5}
}

func (m *Edge) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Edge.Unmarshal(m, b)
}
func (m *Edge) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Edge.Marshal(b, m, deterministic)
}
func (m *Edge) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Edge.Merge(m, src)
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

type Result struct {
	Conclusion           Conclusion                `protobuf:"varint,1,opt,name=conclusion,proto3,enum=adagio.Conclusion" json:"conclusion,omitempty"`
	Metadata             map[string]*MetadataValue `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Output               []byte                    `protobuf:"bytes,3,opt,name=output,proto3" json:"output,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *Result) Reset()         { *m = Result{} }
func (m *Result) String() string { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()    {}
func (*Result) Descriptor() ([]byte, []int) {
	return fileDescriptor_5eb97351c0f66fbe, []int{6}
}

func (m *Result) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Result.Unmarshal(m, b)
}
func (m *Result) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Result.Marshal(b, m, deterministic)
}
func (m *Result) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Result.Merge(m, src)
}
func (m *Result) XXX_Size() int {
	return xxx_messageInfo_Result.Size(m)
}
func (m *Result) XXX_DiscardUnknown() {
	xxx_messageInfo_Result.DiscardUnknown(m)
}

var xxx_messageInfo_Result proto.InternalMessageInfo

func (m *Result) GetConclusion() Conclusion {
	if m != nil {
		return m.Conclusion
	}
	return Conclusion_NONE
}

func (m *Result) GetMetadata() map[string]*MetadataValue {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *Result) GetOutput() []byte {
	if m != nil {
		return m.Output
	}
	return nil
}

func init() {
	proto.RegisterEnum("adagio.Conclusion", Conclusion_name, Conclusion_value)
	proto.RegisterEnum("adagio.Event_Type", Event_Type_name, Event_Type_value)
	proto.RegisterEnum("adagio.Node_Status", Node_Status_name, Node_Status_value)
	proto.RegisterType((*Run)(nil), "adagio.Run")
	proto.RegisterType((*Event)(nil), "adagio.Event")
	proto.RegisterType((*GraphSpec)(nil), "adagio.GraphSpec")
	proto.RegisterType((*MetadataValue)(nil), "adagio.MetadataValue")
	proto.RegisterType((*Node)(nil), "adagio.Node")
	proto.RegisterMapType((map[string][]byte)(nil), "adagio.Node.InputsEntry")
	proto.RegisterType((*Node_Spec)(nil), "adagio.Node.Spec")
	proto.RegisterMapType((map[string]*MetadataValue)(nil), "adagio.Node.Spec.MetadataEntry")
	proto.RegisterType((*Edge)(nil), "adagio.Edge")
	proto.RegisterType((*Result)(nil), "adagio.Result")
	proto.RegisterMapType((map[string]*MetadataValue)(nil), "adagio.Result.MetadataEntry")
}

func init() { proto.RegisterFile("pkg/adagio/adagio.proto", fileDescriptor_5eb97351c0f66fbe) }

var fileDescriptor_5eb97351c0f66fbe = []byte{
	// 689 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x54, 0xdd, 0x4e, 0xdb, 0x4a,
	0x10, 0xc6, 0x89, 0x63, 0xc8, 0x04, 0x90, 0xcf, 0x1e, 0xce, 0xa9, 0x85, 0xa8, 0x88, 0x2c, 0x15,
	0xa2, 0xa2, 0x86, 0x2a, 0xbd, 0xa1, 0x3f, 0x17, 0x75, 0x83, 0x8b, 0x22, 0x81, 0x53, 0x6d, 0x4c,
	0xff, 0x6e, 0xd0, 0x62, 0x6f, 0x8d, 0x05, 0xb1, 0x2d, 0x7b, 0x8d, 0x14, 0xa9, 0x4f, 0xd1, 0x37,
	0xea, 0x45, 0x5f, 0xa4, 0x4f, 0x52, 0xed, 0x8f, 0x83, 0x43, 0x41, 0xea, 0x4d, 0xaf, 0xe2, 0x99,
	0xef, 0xdb, 0xd9, 0x99, 0x6f, 0xbe, 0x2c, 0x3c, 0xc8, 0x2e, 0xa3, 0x7d, 0x12, 0x92, 0x28, 0x4e,
	0xd5, 0x4f, 0x3f, 0xcb, 0x53, 0x96, 0x22, 0x43, 0x46, 0xf6, 0x57, 0x68, 0xe2, 0x32, 0x41, 0xeb,
	0xd0, 0x88, 0x43, 0x4b, 0xeb, 0x6a, 0xbd, 0x36, 0x6e, 0xc4, 0x21, 0x7a, 0x08, 0x10, 0xe4, 0x94,
	0x30, 0x1a, 0x9e, 0x11, 0x66, 0x35, 0x44, 0xbe, 0xad, 0x32, 0x0e, 0x43, 0x36, 0xb4, 0x92, 0x34,
	0xa4, 0x85, 0xd5, 0xec, 0x36, 0x7b, 0x9d, 0xc1, 0x6a, 0x5f, 0xd5, 0xf6, 0xd2, 0x90, 0x62, 0x09,
	0x71, 0x0e, 0x0d, 0x23, 0x5a, 0x58, 0xfa, 0x22, 0xc7, 0x0d, 0x23, 0x8a, 0x25, 0x64, 0x7f, 0xd3,
	0xa0, 0xe5, 0x5e, 0xd3, 0x84, 0xa1, 0x1d, 0xd0, 0xd9, 0x2c, 0xa3, 0xa2, 0x85, 0xf5, 0x01, 0x9a,
	0x93, 0x39, 0xd8, 0xf7, 0x67, 0x19, 0xc5, 0x02, 0x47, 0x1b, 0xd0, 0xca, 0xcb, 0x64, 0x74, 0xa8,
	0x7a, 0x92, 0x01, 0x7a, 0x02, 0x2b, 0xfc, 0xd2, 0x49, 0x46, 0x03, 0xab, 0xd9, 0xd5, 0x7a, 0x9d,
	0xc1, 0x3f, 0xf5, 0x96, 0xfa, 0x1c, 0xc0, 0x73, 0x8a, 0xbd, 0x05, 0xba, 0x2f, 0x8b, 0x99, 0x13,
	0xdf, 0xf1, 0xdd, 0x33, 0x1f, 0x3b, 0xde, 0x64, 0xe4, 0x8f, 0xc6, 0x9e, 0xb9, 0x64, 0x7f, 0x84,
	0xf6, 0x51, 0x4e, 0xb2, 0x0b, 0x4e, 0x45, 0xbb, 0xd5, 0xa4, 0x9a, 0x98, 0xe2, 0x8e, 0xb2, 0xb7,
	0xc7, 0x6d, 0xdc, 0x3f, 0xee, 0x2e, 0xac, 0x9d, 0x50, 0x46, 0x42, 0xc2, 0xc8, 0x7b, 0x72, 0x55,
	0x52, 0xf4, 0x3f, 0x18, 0xd7, 0xfc, 0x43, 0x96, 0x6f, 0x63, 0x15, 0xd9, 0xdf, 0x75, 0xd0, 0xf9,
	0x0d, 0xe8, 0x11, 0xe8, 0x05, 0x1f, 0x4a, 0xbb, 0x6f, 0x28, 0x01, 0xa3, 0x3d, 0x30, 0x0a, 0x46,
	0x58, 0x59, 0x08, 0x59, 0xd6, 0x07, 0xff, 0x2e, 0x12, 0x05, 0x84, 0x15, 0x05, 0x0d, 0x00, 0x82,
	0x34, 0x09, 0xae, 0xca, 0x22, 0x4e, 0x13, 0x21, 0x57, 0x4d, 0xf0, 0xe1, 0x1c, 0xc1, 0x35, 0x16,
	0xf7, 0x43, 0xc1, 0x48, 0xae, 0xfc, 0xa0, 0x4b, 0x3f, 0xa8, 0x8c, 0xc3, 0xd0, 0x36, 0x74, 0xbe,
	0xc4, 0x49, 0x5c, 0x5c, 0x48, 0xbc, 0x25, 0x70, 0xa8, 0x52, 0x0e, 0x43, 0x4f, 0xc1, 0x88, 0x93,
	0xac, 0x64, 0x85, 0x65, 0x08, 0x79, 0xac, 0x85, 0x06, 0x47, 0x02, 0x72, 0x13, 0x96, 0xcf, 0xb0,
	0xe2, 0x6d, 0xfe, 0xd0, 0x40, 0x17, 0x1b, 0x40, 0xa0, 0x27, 0x64, 0x4a, 0x95, 0x39, 0xc5, 0x37,
	0xb2, 0x60, 0x39, 0x2f, 0x13, 0x16, 0x4f, 0xa9, 0xf2, 0x41, 0x15, 0xa2, 0x97, 0xb0, 0x32, 0x55,
	0x12, 0x2b, 0x73, 0x6e, 0xff, 0x26, 0x5a, 0xbf, 0x5a, 0x82, 0xbc, 0x71, 0x7e, 0x60, 0x13, 0xdf,
	0xec, 0x47, 0x40, 0xc8, 0x84, 0xe6, 0x25, 0x9d, 0xa9, 0xab, 0xf9, 0x27, 0xda, 0x83, 0x96, 0xd8,
	0x91, 0xb8, 0xb7, 0x33, 0xf8, 0xaf, 0x2a, 0xbe, 0xb0, 0x57, 0x2c, 0x39, 0x2f, 0x1a, 0x07, 0xda,
	0xe6, 0x73, 0xe8, 0xd4, 0xc6, 0xbb, 0xa3, 0xe2, 0x46, 0xbd, 0xe2, 0x6a, 0xed, 0xa8, 0xfd, 0x0a,
	0x0c, 0xb9, 0x3a, 0xd4, 0x81, 0xe5, 0x0f, 0xce, 0xc8, 0x1f, 0x79, 0x47, 0xe6, 0x12, 0x6a, 0x43,
	0x0b, 0xbb, 0xce, 0xe1, 0x27, 0x53, 0xe3, 0x79, 0x7c, 0xea, 0x79, 0x3c, 0xdf, 0x40, 0x6b, 0xd0,
	0x1e, 0x8e, 0x4f, 0xde, 0x1d, 0xbb, 0xbe, 0x7b, 0x68, 0x36, 0xed, 0xd7, 0xa0, 0x73, 0xef, 0x71,
	0x8f, 0x15, 0x69, 0x99, 0x07, 0x95, 0x82, 0x2a, 0x42, 0x5d, 0xe8, 0x84, 0xb4, 0x60, 0x71, 0x42,
	0x18, 0xf7, 0x81, 0xd4, 0xb1, 0x9e, 0xb2, 0x7f, 0x6a, 0x60, 0x60, 0x5a, 0x94, 0x57, 0xec, 0x96,
	0x67, 0xb4, 0x3f, 0xf2, 0xcc, 0x41, 0x6d, 0x15, 0xf2, 0x4f, 0xb1, 0x55, 0x9d, 0x90, 0x55, 0xef,
	0xdb, 0x03, 0x6f, 0x39, 0x2d, 0x59, 0x56, 0x32, 0xe1, 0xce, 0x55, 0xac, 0xa2, 0xbf, 0xb1, 0x9f,
	0xc7, 0x07, 0x00, 0x37, 0xfd, 0xa3, 0x15, 0xd0, 0xbd, 0xb1, 0xe7, 0x9a, 0x4b, 0x5c, 0xda, 0xc9,
	0xe9, 0x70, 0xe8, 0x4e, 0x26, 0xa6, 0xc6, 0xd3, 0x6f, 0x9d, 0xd1, 0xb1, 0xd9, 0xe0, 0xe2, 0xbb,
	0x18, 0x8f, 0xb1, 0xd9, 0x7c, 0xd3, 0xfb, 0xbc, 0x13, 0xc5, 0xec, 0xa2, 0x3c, 0xef, 0x07, 0xe9,
	0x74, 0x3f, 0xa2, 0x69, 0x1e, 0xd1, 0x29, 0x09, 0xaa, 0xe7, 0xf6, 0xe6, 0xe5, 0x3d, 0x37, 0xc4,
	0x9b, 0xfb, 0xec, 0x57, 0x00, 0x00, 0x00, 0xff, 0xff, 0x0c, 0x6b, 0x36, 0x56, 0x8e, 0x05, 0x00,
	0x00,
}
