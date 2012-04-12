// Code generated by protoc-gen-go from "mixer.proto"
// DO NOT EDIT!

package proto

import proto1 "code.google.com/p/goprotobuf/proto"
import "math"

// Reference proto, math & os imports to suppress error if they are not otherwise used.
var _ = proto1.GetString
var _ = math.Inf
var _ error

type Mixer struct {
	Name             *string    `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Version          *string    `protobuf:"bytes,2,req,name=version" json:"version,omitempty"`
	Assets           []*File    `protobuf:"bytes,4,rep,name=assets" json:"assets,omitempty"`
	Features         []*Feature `protobuf:"bytes,5,rep,name=features" json:"features,omitempty"`
	Recipes          []*Recipe  `protobuf:"bytes,6,rep,name=recipes" json:"recipes,omitempty"`
	Rewriters        []*File    `protobuf:"bytes,7,rep,name=rewriters" json:"rewriters,omitempty"`
	Package          *Package   `protobuf:"bytes,8,opt,name=package" json:"package,omitempty"`
	XXX_unrecognized []byte     `json:",omitempty"`
}

func (this *Mixer) Reset()         { *this = Mixer{} }
func (this *Mixer) String() string { return proto1.CompactTextString(this) }

func init() {
}
