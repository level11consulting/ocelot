// Code generated by protoc-gen-go. DO NOT EDIT.
// source: vcshandler.proto

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import timestamp "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type BranchHistory struct {
	Branch               string               `protobuf:"bytes,1,opt,name=branch" json:"branch,omitempty"`
	Hash                 string               `protobuf:"bytes,2,opt,name=hash" json:"hash,omitempty"`
	LastCommitTime       *timestamp.Timestamp `protobuf:"bytes,3,opt,name=lastCommitTime" json:"lastCommitTime,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *BranchHistory) Reset()         { *m = BranchHistory{} }
func (m *BranchHistory) String() string { return proto.CompactTextString(m) }
func (*BranchHistory) ProtoMessage()    {}
func (*BranchHistory) Descriptor() ([]byte, []int) {
	return fileDescriptor_vcshandler_4f52592eda5aaccb, []int{0}
}
func (m *BranchHistory) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BranchHistory.Unmarshal(m, b)
}
func (m *BranchHistory) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BranchHistory.Marshal(b, m, deterministic)
}
func (dst *BranchHistory) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BranchHistory.Merge(dst, src)
}
func (m *BranchHistory) XXX_Size() int {
	return xxx_messageInfo_BranchHistory.Size(m)
}
func (m *BranchHistory) XXX_DiscardUnknown() {
	xxx_messageInfo_BranchHistory.DiscardUnknown(m)
}

var xxx_messageInfo_BranchHistory proto.InternalMessageInfo

func (m *BranchHistory) GetBranch() string {
	if m != nil {
		return m.Branch
	}
	return ""
}

func (m *BranchHistory) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *BranchHistory) GetLastCommitTime() *timestamp.Timestamp {
	if m != nil {
		return m.LastCommitTime
	}
	return nil
}

type Commit struct {
	Hash                 string               `protobuf:"bytes,1,opt,name=hash" json:"hash,omitempty"`
	Message              string               `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
	Date                 *timestamp.Timestamp `protobuf:"bytes,3,opt,name=date" json:"date,omitempty"`
	Author               *User                `protobuf:"bytes,4,opt,name=author" json:"author,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Commit) Reset()         { *m = Commit{} }
func (m *Commit) String() string { return proto.CompactTextString(m) }
func (*Commit) ProtoMessage()    {}
func (*Commit) Descriptor() ([]byte, []int) {
	return fileDescriptor_vcshandler_4f52592eda5aaccb, []int{1}
}
func (m *Commit) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Commit.Unmarshal(m, b)
}
func (m *Commit) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Commit.Marshal(b, m, deterministic)
}
func (dst *Commit) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Commit.Merge(dst, src)
}
func (m *Commit) XXX_Size() int {
	return xxx_messageInfo_Commit.Size(m)
}
func (m *Commit) XXX_DiscardUnknown() {
	xxx_messageInfo_Commit.DiscardUnknown(m)
}

var xxx_messageInfo_Commit proto.InternalMessageInfo

func (m *Commit) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *Commit) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *Commit) GetDate() *timestamp.Timestamp {
	if m != nil {
		return m.Date
	}
	return nil
}

func (m *Commit) GetAuthor() *User {
	if m != nil {
		return m.Author
	}
	return nil
}

type User struct {
	UserName             string   `protobuf:"bytes,1,opt,name=userName" json:"userName,omitempty"`
	DisplayName          string   `protobuf:"bytes,2,opt,name=displayName" json:"displayName,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *User) Reset()         { *m = User{} }
func (m *User) String() string { return proto.CompactTextString(m) }
func (*User) ProtoMessage()    {}
func (*User) Descriptor() ([]byte, []int) {
	return fileDescriptor_vcshandler_4f52592eda5aaccb, []int{2}
}
func (m *User) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_User.Unmarshal(m, b)
}
func (m *User) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_User.Marshal(b, m, deterministic)
}
func (dst *User) XXX_Merge(src proto.Message) {
	xxx_messageInfo_User.Merge(dst, src)
}
func (m *User) XXX_Size() int {
	return xxx_messageInfo_User.Size(m)
}
func (m *User) XXX_DiscardUnknown() {
	xxx_messageInfo_User.DiscardUnknown(m)
}

var xxx_messageInfo_User proto.InternalMessageInfo

func (m *User) GetUserName() string {
	if m != nil {
		return m.UserName
	}
	return ""
}

func (m *User) GetDisplayName() string {
	if m != nil {
		return m.DisplayName
	}
	return ""
}

type Repo struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	AcctRepo             string   `protobuf:"bytes,2,opt,name=acctRepo" json:"acctRepo,omitempty"`
	RepoLink             string   `protobuf:"bytes,3,opt,name=repoLink" json:"repoLink,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Repo) Reset()         { *m = Repo{} }
func (m *Repo) String() string { return proto.CompactTextString(m) }
func (*Repo) ProtoMessage()    {}
func (*Repo) Descriptor() ([]byte, []int) {
	return fileDescriptor_vcshandler_4f52592eda5aaccb, []int{3}
}
func (m *Repo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Repo.Unmarshal(m, b)
}
func (m *Repo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Repo.Marshal(b, m, deterministic)
}
func (dst *Repo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Repo.Merge(dst, src)
}
func (m *Repo) XXX_Size() int {
	return xxx_messageInfo_Repo.Size(m)
}
func (m *Repo) XXX_DiscardUnknown() {
	xxx_messageInfo_Repo.DiscardUnknown(m)
}

var xxx_messageInfo_Repo proto.InternalMessageInfo

func (m *Repo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Repo) GetAcctRepo() string {
	if m != nil {
		return m.AcctRepo
	}
	return ""
}

func (m *Repo) GetRepoLink() string {
	if m != nil {
		return m.RepoLink
	}
	return ""
}

type Push struct {
	Repo                 *Repo     `protobuf:"bytes,1,opt,name=repo" json:"repo,omitempty"`
	User                 *User     `protobuf:"bytes,2,opt,name=user" json:"user,omitempty"`
	HeadCommit           *Commit   `protobuf:"bytes,4,opt,name=headCommit" json:"headCommit,omitempty"`
	Commits              []*Commit `protobuf:"bytes,3,rep,name=commits" json:"commits,omitempty"`
	Branch               string    `protobuf:"bytes,5,opt,name=Branch" json:"Branch,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *Push) Reset()         { *m = Push{} }
func (m *Push) String() string { return proto.CompactTextString(m) }
func (*Push) ProtoMessage()    {}
func (*Push) Descriptor() ([]byte, []int) {
	return fileDescriptor_vcshandler_4f52592eda5aaccb, []int{4}
}
func (m *Push) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Push.Unmarshal(m, b)
}
func (m *Push) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Push.Marshal(b, m, deterministic)
}
func (dst *Push) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Push.Merge(dst, src)
}
func (m *Push) XXX_Size() int {
	return xxx_messageInfo_Push.Size(m)
}
func (m *Push) XXX_DiscardUnknown() {
	xxx_messageInfo_Push.DiscardUnknown(m)
}

var xxx_messageInfo_Push proto.InternalMessageInfo

func (m *Push) GetRepo() *Repo {
	if m != nil {
		return m.Repo
	}
	return nil
}

func (m *Push) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

func (m *Push) GetHeadCommit() *Commit {
	if m != nil {
		return m.HeadCommit
	}
	return nil
}

func (m *Push) GetCommits() []*Commit {
	if m != nil {
		return m.Commits
	}
	return nil
}

func (m *Push) GetBranch() string {
	if m != nil {
		return m.Branch
	}
	return ""
}

type PullRequest struct {
	Description          string    `protobuf:"bytes,1,opt,name=description" json:"description,omitempty"`
	Urls                 *PrUrls   `protobuf:"bytes,2,opt,name=urls" json:"urls,omitempty"`
	Title                string    `protobuf:"bytes,3,opt,name=title" json:"title,omitempty"`
	Source               *HeadData `protobuf:"bytes,4,opt,name=source" json:"source,omitempty"`
	Destination          *HeadData `protobuf:"bytes,5,opt,name=destination" json:"destination,omitempty"`
	Id                   int64     `protobuf:"varint,6,opt,name=id" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *PullRequest) Reset()         { *m = PullRequest{} }
func (m *PullRequest) String() string { return proto.CompactTextString(m) }
func (*PullRequest) ProtoMessage()    {}
func (*PullRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_vcshandler_4f52592eda5aaccb, []int{5}
}
func (m *PullRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PullRequest.Unmarshal(m, b)
}
func (m *PullRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PullRequest.Marshal(b, m, deterministic)
}
func (dst *PullRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PullRequest.Merge(dst, src)
}
func (m *PullRequest) XXX_Size() int {
	return xxx_messageInfo_PullRequest.Size(m)
}
func (m *PullRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PullRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PullRequest proto.InternalMessageInfo

func (m *PullRequest) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *PullRequest) GetUrls() *PrUrls {
	if m != nil {
		return m.Urls
	}
	return nil
}

func (m *PullRequest) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *PullRequest) GetSource() *HeadData {
	if m != nil {
		return m.Source
	}
	return nil
}

func (m *PullRequest) GetDestination() *HeadData {
	if m != nil {
		return m.Destination
	}
	return nil
}

func (m *PullRequest) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

type HeadData struct {
	Branch               string   `protobuf:"bytes,1,opt,name=branch" json:"branch,omitempty"`
	Hash                 string   `protobuf:"bytes,2,opt,name=hash" json:"hash,omitempty"`
	Repo                 *Repo    `protobuf:"bytes,3,opt,name=repo" json:"repo,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HeadData) Reset()         { *m = HeadData{} }
func (m *HeadData) String() string { return proto.CompactTextString(m) }
func (*HeadData) ProtoMessage()    {}
func (*HeadData) Descriptor() ([]byte, []int) {
	return fileDescriptor_vcshandler_4f52592eda5aaccb, []int{6}
}
func (m *HeadData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HeadData.Unmarshal(m, b)
}
func (m *HeadData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HeadData.Marshal(b, m, deterministic)
}
func (dst *HeadData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HeadData.Merge(dst, src)
}
func (m *HeadData) XXX_Size() int {
	return xxx_messageInfo_HeadData.Size(m)
}
func (m *HeadData) XXX_DiscardUnknown() {
	xxx_messageInfo_HeadData.DiscardUnknown(m)
}

var xxx_messageInfo_HeadData proto.InternalMessageInfo

func (m *HeadData) GetBranch() string {
	if m != nil {
		return m.Branch
	}
	return ""
}

func (m *HeadData) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *HeadData) GetRepo() *Repo {
	if m != nil {
		return m.Repo
	}
	return nil
}

type PrUrls struct {
	Commits              string   `protobuf:"bytes,1,opt,name=commits" json:"commits,omitempty"`
	Comments             string   `protobuf:"bytes,2,opt,name=comments" json:"comments,omitempty"`
	Statuses             string   `protobuf:"bytes,3,opt,name=statuses" json:"statuses,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PrUrls) Reset()         { *m = PrUrls{} }
func (m *PrUrls) String() string { return proto.CompactTextString(m) }
func (*PrUrls) ProtoMessage()    {}
func (*PrUrls) Descriptor() ([]byte, []int) {
	return fileDescriptor_vcshandler_4f52592eda5aaccb, []int{7}
}
func (m *PrUrls) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PrUrls.Unmarshal(m, b)
}
func (m *PrUrls) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PrUrls.Marshal(b, m, deterministic)
}
func (dst *PrUrls) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PrUrls.Merge(dst, src)
}
func (m *PrUrls) XXX_Size() int {
	return xxx_messageInfo_PrUrls.Size(m)
}
func (m *PrUrls) XXX_DiscardUnknown() {
	xxx_messageInfo_PrUrls.DiscardUnknown(m)
}

var xxx_messageInfo_PrUrls proto.InternalMessageInfo

func (m *PrUrls) GetCommits() string {
	if m != nil {
		return m.Commits
	}
	return ""
}

func (m *PrUrls) GetComments() string {
	if m != nil {
		return m.Comments
	}
	return ""
}

func (m *PrUrls) GetStatuses() string {
	if m != nil {
		return m.Statuses
	}
	return ""
}

func init() {
	proto.RegisterType((*BranchHistory)(nil), "models.BranchHistory")
	proto.RegisterType((*Commit)(nil), "models.Commit")
	proto.RegisterType((*User)(nil), "models.User")
	proto.RegisterType((*Repo)(nil), "models.Repo")
	proto.RegisterType((*Push)(nil), "models.Push")
	proto.RegisterType((*PullRequest)(nil), "models.PullRequest")
	proto.RegisterType((*HeadData)(nil), "models.HeadData")
	proto.RegisterType((*PrUrls)(nil), "models.PrUrls")
}

func init() { proto.RegisterFile("vcshandler.proto", fileDescriptor_vcshandler_4f52592eda5aaccb) }

var fileDescriptor_vcshandler_4f52592eda5aaccb = []byte{
	// 517 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0xcd, 0x6a, 0xdc, 0x30,
	0x10, 0xc6, 0xbb, 0x8e, 0xb3, 0x99, 0x6d, 0x97, 0x20, 0x4a, 0x31, 0x7b, 0xa9, 0x31, 0x3d, 0xec,
	0xc9, 0x81, 0xf4, 0x0d, 0xb6, 0x39, 0xe4, 0x50, 0xca, 0x22, 0x12, 0x28, 0xb9, 0x69, 0xed, 0xe9,
	0x5a, 0xd4, 0xb6, 0x5c, 0x8d, 0x5c, 0xc8, 0xa9, 0x2f, 0xd0, 0x47, 0xea, 0x7b, 0xf4, 0x75, 0x8a,
	0x24, 0xcb, 0x35, 0xa1, 0x81, 0xf6, 0xa6, 0x6f, 0xbe, 0x4f, 0xf3, 0xf3, 0x69, 0x04, 0x97, 0xdf,
	0x4a, 0xaa, 0x45, 0x57, 0x35, 0xa8, 0x8b, 0x5e, 0x2b, 0xa3, 0x58, 0xd2, 0xaa, 0x0a, 0x1b, 0xda,
	0xbe, 0x39, 0x29, 0x75, 0x6a, 0xf0, 0xca, 0x45, 0x8f, 0xc3, 0xe7, 0x2b, 0x23, 0x5b, 0x24, 0x23,
	0xda, 0xde, 0x0b, 0xf3, 0xef, 0xf0, 0x72, 0xaf, 0x45, 0x57, 0xd6, 0xb7, 0x92, 0x8c, 0xd2, 0x8f,
	0xec, 0x35, 0x24, 0x47, 0x17, 0x48, 0xa3, 0x2c, 0xda, 0x5d, 0xf0, 0x11, 0x31, 0x06, 0x71, 0x2d,
	0xa8, 0x4e, 0x17, 0x2e, 0xea, 0xce, 0x6c, 0x0f, 0x9b, 0x46, 0x90, 0x79, 0xaf, 0xda, 0x56, 0x9a,
	0x3b, 0xd9, 0x62, 0xba, 0xcc, 0xa2, 0xdd, 0xfa, 0x7a, 0x5b, 0xf8, 0xb2, 0x45, 0x28, 0x5b, 0xdc,
	0x85, 0xb2, 0xfc, 0xc9, 0x8d, 0xfc, 0x47, 0x04, 0x89, 0x87, 0x53, 0x89, 0x68, 0x56, 0x22, 0x85,
	0xf3, 0x16, 0x89, 0xc4, 0x09, 0xc7, 0xca, 0x01, 0xb2, 0x02, 0xe2, 0x4a, 0x98, 0x7f, 0x29, 0xe9,
	0x74, 0xec, 0x2d, 0x24, 0x62, 0x30, 0xb5, 0xd2, 0x69, 0xec, 0x6e, 0xbc, 0x28, 0xbc, 0x47, 0xc5,
	0x3d, 0xa1, 0xe6, 0x23, 0x97, 0xdf, 0x40, 0x6c, 0x31, 0xdb, 0xc2, 0x6a, 0x20, 0xd4, 0x1f, 0x45,
	0x8b, 0x63, 0x3f, 0x13, 0x66, 0x19, 0xac, 0x2b, 0x49, 0x7d, 0x23, 0x1e, 0x1d, 0xed, 0xfb, 0x9a,
	0x87, 0x72, 0x0e, 0x31, 0xc7, 0x5e, 0xd9, 0x89, 0xba, 0x3f, 0x19, 0xdc, 0xd9, 0x66, 0x16, 0x65,
	0x69, 0x2c, 0x3f, 0x5e, 0x9d, 0xb0, 0xe5, 0x34, 0xf6, 0xea, 0x83, 0xec, 0xbe, 0xb8, 0xb9, 0x2e,
	0xf8, 0x84, 0xf3, 0x9f, 0x11, 0xc4, 0x87, 0x81, 0x6a, 0x96, 0x41, 0x6c, 0x83, 0x2e, 0xe9, 0x6c,
	0x0c, 0x9b, 0x80, 0x3b, 0xc6, 0x2a, 0x6c, 0xb3, 0x2e, 0xfd, 0xd3, 0x41, 0x1d, 0xc3, 0x0a, 0x80,
	0x1a, 0x45, 0xe5, 0x8d, 0x1f, 0x0d, 0xd9, 0x04, 0x9d, 0x8f, 0xf2, 0x99, 0x82, 0xed, 0xe0, 0xbc,
	0x74, 0x27, 0x4a, 0x97, 0xd9, 0xf2, 0x2f, 0xe2, 0x40, 0xdb, 0xfd, 0xf1, 0x0b, 0x95, 0x9e, 0xf9,
	0xfd, 0xf1, 0x28, 0xff, 0x15, 0xc1, 0xfa, 0x30, 0x34, 0x0d, 0xc7, 0xaf, 0x03, 0x92, 0x71, 0x26,
	0x22, 0x95, 0x5a, 0xf6, 0x46, 0xaa, 0x6e, 0x74, 0x68, 0x1e, 0x62, 0x39, 0xc4, 0x83, 0x6e, 0x68,
	0x9c, 0x62, 0x2a, 0x78, 0xd0, 0xf7, 0xba, 0x21, 0xee, 0x38, 0xf6, 0x0a, 0xce, 0x8c, 0x34, 0x0d,
	0x8e, 0x6e, 0x79, 0xc0, 0x76, 0x90, 0x90, 0x1a, 0x74, 0x89, 0xe3, 0x64, 0x97, 0xe1, 0xee, 0x2d,
	0x8a, 0xea, 0x46, 0x18, 0xc1, 0x47, 0x9e, 0x5d, 0xbb, 0x2e, 0x8c, 0xec, 0x84, 0xeb, 0xe2, 0xec,
	0x19, 0xf9, 0x5c, 0xc4, 0x36, 0xb0, 0x90, 0x55, 0x9a, 0x64, 0xd1, 0x6e, 0xc9, 0x17, 0xb2, 0xca,
	0x3f, 0xc1, 0x2a, 0x08, 0xff, 0xeb, 0xf7, 0x84, 0x77, 0x5c, 0x3e, 0xf7, 0x8e, 0xf9, 0x03, 0x24,
	0x7e, 0x5a, 0xfb, 0x0d, 0x82, 0xff, 0x3e, 0xf1, 0xe4, 0xf7, 0x16, 0x56, 0xf6, 0x88, 0x9d, 0xa1,
	0xb0, 0x4e, 0x01, 0x5b, 0x8e, 0x8c, 0x30, 0x03, 0x21, 0x85, 0x75, 0x0a, 0x78, 0x1f, 0x3f, 0x2c,
	0xfa, 0xe3, 0x31, 0x71, 0xdf, 0xe5, 0xdd, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0x9f, 0x31, 0x2a,
	0xce, 0x42, 0x04, 0x00, 0x00,
}
