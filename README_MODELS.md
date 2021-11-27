# Declaring Models
[![Go
Reference](https://pkg.go.dev/badge/github.com/hashicorp/go-dbw.svg)](https://pkg.go.dev/github.com/hashicorp/go-dbw)

Models are structs with basic Go types, or custom types implementing [Scanner](https://pkg.go.dev/database/sql#Scanner) and
[Valuer](https://pkg.go.dev/database/sql/driver#Valuer) interfaces. Currently,
Gorm V2 [Field Tags](https://gorm.io/docs/models.html#Fields-Tags) are supported when declaring models. 

Simple example:

```go
type TestUser struct {
    PublicId    string      `json:"public_id,omitempty" gorm:"primaryKey;default:null"`
	CreateTime  *time.Time  `json:"create_time,omitempty" gorm:"default:CURRENT_TIMESTAMP"`
	UpdateTime  *time.Time  `json:"create_time,omitempty" gorm:"default:CURRENT_TIMESTAMP"`
	Name        string      `json:"name,omitempty" gorm:"default:null"`
	Email 		string 		`json:"name,omitempty" gorm:"default:null"`
	PhoneNumber	string 		`json:"name,omitempty" gorm:"default:null"`
	Version     uint32      `json:"version,omitempty" gorm:"default:null"`
}

```
A more complicated example that uses an embedded protobuf:
```go
type TestUser struct {
	*StoreTestUser
}

// TestUser model
type StoreTestUser struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @inject_tag: gorm:"primaryKey;default:null"
	PublicId string `protobuf:"bytes,4,opt,name=public_id,json=publicId,proto3" json:"public_id,omitempty" gorm:"primaryKey;default:null"`

    // @inject_tag: `gorm:"default:CURRENT_TIMESTAMP"`
	CreateTime *Timestamp `protobuf:"bytes,2,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty" gorm:"default:CURRENT_TIMESTAMP"`

    // @inject_tag: `gorm:"default:CURRENT_TIMESTAMP"`
	UpdateTime *Timestamp `protobuf:"bytes,3,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty" gorm:"default:CURRENT_TIMESTAMP"`

	// @inject_tag: `gorm:"default:null"`
    Name string `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty" gorm:"default:null"`
    
	// @inject_tag: `gorm:"default:null"`
	Version uint32 `protobuf:"varint,8,opt,name=version,proto3" json:"version,omitempty" gorm:"default:null"`
}

```
