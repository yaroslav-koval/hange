package errmapper

import (
	"os"

	"github.com/yaroslav-koval/hange/domain/fileprovider"
)

func NewOSFileErrMapper() FileErrorMapper {
	mapping := map[error]error{
		os.ErrNotExist:   fileprovider.ErrNotExist,
		os.ErrPermission: fileprovider.ErrPermission,
	}

	return &osFileErrMapper{
		mapping: mapping,
	}
}

type osFileErrMapper struct {
	mapping map[error]error
}

func (o *osFileErrMapper) Map(err error) error {
	pkgErr, ok := o.mapping[err]
	if ok {
		return pkgErr
	}

	return err
}
