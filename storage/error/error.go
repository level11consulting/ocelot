package storage_error

import (
	"fmt"

	"github.com/level11consulting/ocelot/models/pb"
)

var (
	BUILD_SUM_404    = "no build summary found for %s"
	STAGE_REASON_404 = "no stages found for %s"
	BUILD_OUT_404    = "no build output found for %s"
	CRED_404         = "no credential found for %s %s"
)

func BuildSumNotFound(id string) *ErrNotFound {
	return &ErrNotFound{fmt.Sprintf(BUILD_SUM_404, id)}
}

func BuildOutNotFound(id string) *ErrNotFound {
	return &ErrNotFound{fmt.Sprintf(BUILD_OUT_404, id)}
}

func StagesNotFound(id string) *ErrNotFound {
	return &ErrNotFound{fmt.Sprintf(STAGE_REASON_404, id)}
}

func CredNotFound(account string, repoType string) *ErrNotFound {
	return &ErrNotFound{fmt.Sprintf(CRED_404, account, repoType)}
}

type ErrNotFound struct {
	Msg string
}

func (e *ErrNotFound) Error() string {
	return e.Msg
}

func MultipleVCSTypes(account string, types []pb.SubCredType) *ErrMultipleVCSTypes {
	return &ErrMultipleVCSTypes{account: account, types: types}
}

type ErrMultipleVCSTypes struct {
	account string
	types   []pb.SubCredType
}

func (e *ErrMultipleVCSTypes) Error() string {
	var stringTypes []string
	for _, sct := range e.types {
		stringTypes = append(stringTypes, sct.String())
	}
	return fmt.Sprintf("there are multiple vcs types to the account %s: %s", e.account, stringTypes)
}
