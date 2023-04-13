package parser

import (
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
)

type IDevfileCtx interface {
	SetDevfileJSONSchema() error
	ValidateDevfileSchema() error
	GetFs() filesystem.Filesystem
	SetDevfileAPIVersion() error
	GetApiVersion() string
	IsApiVersionSupported() bool
	populateDevfile() (err error)
	Populate() (err error)
	PopulateFromURL() (err error)
	PopulateFromRaw() (err error)
	Validate() error
	GetAbsPath() string
	GetURL() string
	GetToken() string
	SetAbsPath() (err error)
	GetConvertUriToInlined() bool
	SetConvertUriToInlined(value bool)
	SetDevfileContent() error
	SetDevfileContentFromBytes(data []byte) error
	GetDevfileContent() []byte
}
