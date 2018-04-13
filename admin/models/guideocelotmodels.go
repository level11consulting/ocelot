package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"


	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"google.golang.org/grpc"
)

//OcyCredder is an interface for interacting with credentials in Ocelot
type OcyCredder interface {
	SetSecret(string)
	UnmarshalAdditionalFields(fields []byte) error
	CreateAdditionalFields() ([]byte, error)
	GetClientSecret() string
	GetAcctName() string
	GetIdentifier() string
	GetSubType() SubCredType
	ValidateForInsert() *ValidationErr
}

func Invalidate(reason string) *ValidationErr {
	return &ValidationErr{msg: reason}
}

type ValidationErr struct {
	msg string
}

func (v *ValidationErr) Error() string {
	return v.msg
}

func NewRepoCreds() *RepoCreds {
	return &RepoCreds{}
}


func (m *RepoCreds) SetSecret(secret string) {
	m.Password = secret
}

func (m *RepoCreds) GetClientSecret() string {
	return m.Password
}


func (m *RepoCreds) CreateAdditionalFields() ([]byte, error) {
	fields := make(map[string]string)
	fields["username"] = m.Username
	fields["url"] = m.RepoUrl
	bytes, err := json.Marshal(fields)
	return bytes, err
}

func (m *RepoCreds) UnmarshalAdditionalFields(fields []byte) error {
	unmarshaled := make(map[string]string)
	if err := json.Unmarshal(fields, &unmarshaled); err != nil {
		return err
	}
	var ok bool
	if m.RepoUrl, ok = unmarshaled["url"]; !ok {
		return errors.New(fmt.Sprintf("repo url was not in field map, map is %v", unmarshaled))
	}
	if m.Username, ok = unmarshaled["username"]; !ok {
		return errors.New(fmt.Sprintf("username was not in field map, map is %v", unmarshaled))
	}
	return nil
}

func (m *RepoCreds) ValidateForInsert() *ValidationErr {
	errr := validateCommonFieldsForInsert(m)
	if m.RepoUrl == "" {
		errr = append(errr, "repoUrl is required")
	}
	if m.Username == "" {
		errr = append(errr, "username is required")
	}
	if len(errr) != 0 {
		return Invalidate(strings.Join(errr, "\n"))
	}
	return nil
}


func NewVCSCreds() *VCSCreds {
	return &VCSCreds{}
}


func (m *VCSCreds) CreateAdditionalFields() ([]byte, error) {
	fields := make(map[string]string)
	fields["tokenUrl"] = m.TokenURL
	fields["clientId"] = m.ClientId
	bytes, err := json.Marshal(fields)
	return bytes, err
}

func (m *VCSCreds) UnmarshalAdditionalFields(fields []byte) error {
	unmarshaled := make(map[string]string)
	if err := json.Unmarshal(fields, &unmarshaled); err != nil {
		return err
	}
	var ok bool
	if m.TokenURL, ok = unmarshaled["tokenUrl"]; !ok {
		return errors.New(fmt.Sprintf("token url was not in field map, map is %v", unmarshaled))
	}
	if m.ClientId, ok = unmarshaled["clientId"]; !ok {
		return errors.New(fmt.Sprintf("client id was not in field map, map is %v", unmarshaled))
	}
	return nil
}

func (m *VCSCreds) SetSecret(sec string) {
	m.ClientSecret = sec
}

// identifier for vcs creds will always be "<BITBUCKET|GITHUB|..>/<ACCTNAME>"
func (m *VCSCreds) BuildIdentifier() string {
	// can ignore error here, because VcsCreds will always have subtype in VCS
	identifier, _ := CreateVCSIdentifier(m.SubType, m.AcctName)
	return identifier
}

func (m *VCSCreds) ValidateForInsert() *ValidationErr {
	errr := validateCommonFieldsForInsert(m)
	if m.ClientId == "" {
		errr = append(errr, "oauth client id is required")
	}
	if m.TokenURL == "" {
		errr = append(errr, "oauth token url is required")
	}
	if len(errr) != 0 {
		return Invalidate(strings.Join(errr, "\n"))
	}
	return nil
}


func NewK8sCreds() *K8SCreds {
	return &K8SCreds{}
}

func (m *K8SCreds) GetClientSecret() string {
	return m.K8SContents
}

func (m *K8SCreds) SetAcctNameAndType(name, typ string) {
	m.AcctName = name
	// no type here! mua ha ha. GetType() returns a dummy
}

func (m *K8SCreds) SetSecret(str string) {
	m.K8SContents = str
}


func (m *K8SCreds) CreateAdditionalFields() ([]byte, error) {
	return []byte("{}"), nil
}

func (m *K8SCreds) UnmarshalAdditionalFields(fields []byte) error {
	return nil
}


func (m *K8SCreds) ValidateForInsert() *ValidationErr {
	errr := validateCommonFieldsForInsert(m)
	if len(errr) != 0 {
		return Invalidate(strings.Join(errr, "\n"))
	}
	return nil
}


func validateCommonFieldsForInsert(credder OcyCredder) (errors []string) {
	if credder.GetIdentifier() == "" {
		errors = append(errors, "identifier is required, creds need a unique name to identify by")
	}
	if credder.GetAcctName() == "" {
		errors = append(errors, "account name is required")
	}
	if credder.GetClientSecret() == "" {
		errors = append(errors, "client secret is required")
	}
	if credder.GetSubType() == SubCredType_NIL_SCT {
		errors = append(errors, "subtype not instantiated")
	}
	return
}


// wrapper interface around models.BuildRuntimeInfo
type BuildRuntime interface {
	GetDone() bool
	GetIp() string
	GetGrpcPort() string
	GetHash() string
	CreateBuildClient() (protobuf.BuildClient, error)
}


// CreateBuildClient dials the grpc server at the werker endpoints
func (m *BuildRuntimeInfo) CreateBuildClient() (protobuf.BuildClient, error) {
	//TODO: this is insecure
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err :=  grpc.Dial(m.Ip + ":" + m.GrpcPort, opts...)
	if err != nil {
		return nil, err
	}
	return protobuf.NewBuildClient(conn), nil
}

func (pr *PollRequest) Validate() error {
	pr.Cron = strings.TrimSpace(pr.Cron)
	//todo: add validating acct/repo
	return nil
}