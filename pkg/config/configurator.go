package config

type FileType string

const FileTypeYaml FileType = "yaml"

type Configurator interface {
	WriteField(field string, value any) error
	ReadField(field string) any
}
