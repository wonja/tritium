// Code generated by protoc-gen-go.
// source: tritium.proto
// DO NOT EDIT!

package proto

import proto1 "code.google.com/p/goprotobuf/proto"
import json "encoding/json"
import math "math"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto1.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type Transform struct {
	Objects          []*ScriptObject `protobuf:"bytes,1,rep,name=objects" json:"objects,omitempty"`
	Pkg              *Package        `protobuf:"bytes,2,req,name=pkg" json:"pkg,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (this *Transform) Reset()         { *this = Transform{} }
func (this *Transform) String() string { return proto1.CompactTextString(this) }
func (*Transform) ProtoMessage()       {}

func (this *Transform) GetPkg() *Package {
	if this != nil {
		return this.Pkg
	}
	return nil
}

func init() {
}
