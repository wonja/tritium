// Code generated by protoc-gen-go.
// source: rewriter.proto
// DO NOT EDIT!

package proto

import proto1 "code.google.com/p/goprotobuf/proto"
import json "encoding/json"
import math "math"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto1.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type Rewriter struct {
	Name             *string `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Templates        []*File `protobuf:"bytes,2,rep,name=templates" json:"templates,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (this *Rewriter) Reset()         { *this = Rewriter{} }
func (this *Rewriter) String() string { return proto1.CompactTextString(this) }
func (*Rewriter) ProtoMessage()       {}

func (this *Rewriter) GetName() string {
	if this != nil && this.Name != nil {
		return *this.Name
	}
	return ""
}

func init() {
}
