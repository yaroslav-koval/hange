package config

import (
	"bytes"

	"github.com/spf13/viper"
)

// ReadFieldFromBytes makes copy of a file and keeps original bytes immutable
func ReadFieldFromBytes(file []byte, fileType FileType, fieldPath string) (any, error) {
	r := viper.New()
	r.SetConfigType(string(fileType))

	cpFile := make([]byte, len(file))
	copy(cpFile, file)

	if err := r.ReadConfig(bytes.NewBuffer(cpFile)); err != nil {
		return nil, err
	}

	return r.Get(fieldPath), nil
}
