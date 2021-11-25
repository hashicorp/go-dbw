// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: dbtest/storage/v1/dbtest.proto

// define a test proto package for the internal/db package.  These protos
// are only used for unit tests and are not part of the rest of the boundary
// domain model

package dbtest

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Timestamp for storage messages.  We've defined a new local type wrapper
// of google.protobuf.Timestamp so we can implement sql.Scanner and sql.Valuer
// interfaces.  See:
// https://golang.org/pkg/database/sql/#Scanner
// https://golang.org/pkg/database/sql/driver/#Valuer
type Timestamp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Timestamp) Reset() {
	*x = Timestamp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbtest_storage_v1_dbtest_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Timestamp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Timestamp) ProtoMessage() {}

func (x *Timestamp) ProtoReflect() protoreflect.Message {
	mi := &file_dbtest_storage_v1_dbtest_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Timestamp.ProtoReflect.Descriptor instead.
func (*Timestamp) Descriptor() ([]byte, []int) {
	return file_dbtest_storage_v1_dbtest_proto_rawDescGZIP(), []int{0}
}

func (x *Timestamp) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

// TestUser model
type StoreTestUser struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// public_id is the used to access the user via an API
	// @inject_tag: gorm:"primary_key;default:null"
	PublicId string `protobuf:"bytes,4,opt,name=public_id,json=publicId,proto3" json:"public_id,omitempty" gorm:"primary_key;default:null"`
	// create_time from the RDBMS
	// @inject_tag: `gorm:"default:CURRENT_TIMESTAMP"`
	CreateTime *Timestamp `protobuf:"bytes,2,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty" gorm:"default:CURRENT_TIMESTAMP"`
	// update_time from the RDBMS
	// @inject_tag: `gorm:"default:CURRENT_TIMESTAMP"`
	UpdateTime *Timestamp `protobuf:"bytes,3,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty" gorm:"default:CURRENT_TIMESTAMP"`
	// name is the optional friendly name used to
	// access the user via an API
	// @inject_tag: `gorm:"default:null"`
	Name string `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty" gorm:"default:null"`
	// @inject_tag: `gorm:"default:null"`
	PhoneNumber string `protobuf:"bytes,6,opt,name=phone_number,json=phoneNumber,proto3" json:"phone_number,omitempty" gorm:"default:null"`
	// @inject_tag: `gorm:"default:null"`
	Email string `protobuf:"bytes,7,opt,name=email,proto3" json:"email,omitempty" gorm:"default:null"`
	// @inject_tag: `gorm:"default:null"`
	Version uint32 `protobuf:"varint,8,opt,name=version,proto3" json:"version,omitempty" gorm:"default:null"`
}

func (x *StoreTestUser) Reset() {
	*x = StoreTestUser{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbtest_storage_v1_dbtest_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreTestUser) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreTestUser) ProtoMessage() {}

func (x *StoreTestUser) ProtoReflect() protoreflect.Message {
	mi := &file_dbtest_storage_v1_dbtest_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreTestUser.ProtoReflect.Descriptor instead.
func (*StoreTestUser) Descriptor() ([]byte, []int) {
	return file_dbtest_storage_v1_dbtest_proto_rawDescGZIP(), []int{1}
}

func (x *StoreTestUser) GetPublicId() string {
	if x != nil {
		return x.PublicId
	}
	return ""
}

func (x *StoreTestUser) GetCreateTime() *Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *StoreTestUser) GetUpdateTime() *Timestamp {
	if x != nil {
		return x.UpdateTime
	}
	return nil
}

func (x *StoreTestUser) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *StoreTestUser) GetPhoneNumber() string {
	if x != nil {
		return x.PhoneNumber
	}
	return ""
}

func (x *StoreTestUser) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *StoreTestUser) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

// TestCar car model
type StoreTestCar struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// public_id is the used to access the car via an API
	// @inject_tag: gorm:"primary_key;default:null"
	PublicId string `protobuf:"bytes,4,opt,name=public_id,json=publicId,proto3" json:"public_id,omitempty" gorm:"primary_key;default:null"`
	// create_time from the RDBMS
	// @inject_tag: `gorm:"default:CURRENT_TIMESTAMP"`
	CreateTime *Timestamp `protobuf:"bytes,2,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty" gorm:"default:CURRENT_TIMESTAMP"`
	// update_time from the RDBMS
	// @inject_tag: `gorm:"default:CURRENT_TIMESTAMP"`
	UpdateTime *Timestamp `protobuf:"bytes,3,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty" gorm:"default:CURRENT_TIMESTAMP"`
	// name is the optional friendly name used to
	// access the Scope via an API
	// @inject_tag: `gorm:"default:null"`
	Name string `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty" gorm:"default:null"`
	// @inject_tag: `gorm:"default:null"`
	Model string `protobuf:"bytes,6,opt,name=model,proto3" json:"model,omitempty" gorm:"default:null"`
	// @inject_tag: `gorm:"default:null"`
	Mpg int32 `protobuf:"varint,7,opt,name=mpg,proto3" json:"mpg,omitempty" gorm:"default:null"`
}

func (x *StoreTestCar) Reset() {
	*x = StoreTestCar{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dbtest_storage_v1_dbtest_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreTestCar) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreTestCar) ProtoMessage() {}

func (x *StoreTestCar) ProtoReflect() protoreflect.Message {
	mi := &file_dbtest_storage_v1_dbtest_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreTestCar.ProtoReflect.Descriptor instead.
func (*StoreTestCar) Descriptor() ([]byte, []int) {
	return file_dbtest_storage_v1_dbtest_proto_rawDescGZIP(), []int{2}
}

func (x *StoreTestCar) GetPublicId() string {
	if x != nil {
		return x.PublicId
	}
	return ""
}

func (x *StoreTestCar) GetCreateTime() *Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *StoreTestCar) GetUpdateTime() *Timestamp {
	if x != nil {
		return x.UpdateTime
	}
	return nil
}

func (x *StoreTestCar) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *StoreTestCar) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

func (x *StoreTestCar) GetMpg() int32 {
	if x != nil {
		return x.Mpg
	}
	return 0
}

var File_dbtest_storage_v1_dbtest_proto protoreflect.FileDescriptor

var file_dbtest_storage_v1_dbtest_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x64, 0x62, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x2f, 0x76, 0x31, 0x2f, 0x64, 0x62, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x11, 0x64, 0x62, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x45, 0x0a, 0x09, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22, 0x91, 0x02, 0x0a, 0x0d,
	0x53, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x65, 0x73, 0x74, 0x55, 0x73, 0x65, 0x72, 0x12, 0x1b, 0x0a,
	0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x49, 0x64, 0x12, 0x3d, 0x0a, 0x0b, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x64, 0x62, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x2e, 0x76, 0x31, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x3d, 0x0a, 0x0b, 0x75, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x64, 0x62, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x75, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0c,
	0x70, 0x68, 0x6f, 0x6e, 0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x70, 0x68, 0x6f, 0x6e, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12,
	0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x65, 0x6d, 0x61, 0x69, 0x6c, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22,
	0xe5, 0x01, 0x0a, 0x0c, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x65, 0x73, 0x74, 0x43, 0x61, 0x72,
	0x12, 0x1b, 0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x49, 0x64, 0x12, 0x3d, 0x0a,
	0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x64, 0x62, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x74, 0x6f, 0x72,
	0x61, 0x67, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x3d, 0x0a, 0x0b,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1c, 0x2e, 0x64, 0x62, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52,
	0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x70, 0x67, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x03, 0x6d, 0x70, 0x67, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x68, 0x61, 0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x2f,
	0x64, 0x62, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x64, 0x62, 0x74, 0x65,
	0x73, 0x74, 0x3b, 0x64, 0x62, 0x74, 0x65, 0x73, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_dbtest_storage_v1_dbtest_proto_rawDescOnce sync.Once
	file_dbtest_storage_v1_dbtest_proto_rawDescData = file_dbtest_storage_v1_dbtest_proto_rawDesc
)

func file_dbtest_storage_v1_dbtest_proto_rawDescGZIP() []byte {
	file_dbtest_storage_v1_dbtest_proto_rawDescOnce.Do(func() {
		file_dbtest_storage_v1_dbtest_proto_rawDescData = protoimpl.X.CompressGZIP(file_dbtest_storage_v1_dbtest_proto_rawDescData)
	})
	return file_dbtest_storage_v1_dbtest_proto_rawDescData
}

var file_dbtest_storage_v1_dbtest_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_dbtest_storage_v1_dbtest_proto_goTypes = []interface{}{
	(*Timestamp)(nil),             // 0: dbtest.storage.v1.Timestamp
	(*StoreTestUser)(nil),         // 1: dbtest.storage.v1.StoreTestUser
	(*StoreTestCar)(nil),          // 2: dbtest.storage.v1.StoreTestCar
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
}
var file_dbtest_storage_v1_dbtest_proto_depIdxs = []int32{
	3, // 0: dbtest.storage.v1.Timestamp.timestamp:type_name -> google.protobuf.Timestamp
	0, // 1: dbtest.storage.v1.StoreTestUser.create_time:type_name -> dbtest.storage.v1.Timestamp
	0, // 2: dbtest.storage.v1.StoreTestUser.update_time:type_name -> dbtest.storage.v1.Timestamp
	0, // 3: dbtest.storage.v1.StoreTestCar.create_time:type_name -> dbtest.storage.v1.Timestamp
	0, // 4: dbtest.storage.v1.StoreTestCar.update_time:type_name -> dbtest.storage.v1.Timestamp
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_dbtest_storage_v1_dbtest_proto_init() }
func file_dbtest_storage_v1_dbtest_proto_init() {
	if File_dbtest_storage_v1_dbtest_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_dbtest_storage_v1_dbtest_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Timestamp); i {
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
		file_dbtest_storage_v1_dbtest_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreTestUser); i {
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
		file_dbtest_storage_v1_dbtest_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreTestCar); i {
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
			RawDescriptor: file_dbtest_storage_v1_dbtest_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_dbtest_storage_v1_dbtest_proto_goTypes,
		DependencyIndexes: file_dbtest_storage_v1_dbtest_proto_depIdxs,
		MessageInfos:      file_dbtest_storage_v1_dbtest_proto_msgTypes,
	}.Build()
	File_dbtest_storage_v1_dbtest_proto = out.File
	file_dbtest_storage_v1_dbtest_proto_rawDesc = nil
	file_dbtest_storage_v1_dbtest_proto_goTypes = nil
	file_dbtest_storage_v1_dbtest_proto_depIdxs = nil
}
