package errmapper

type FileErrorMapper interface {
	Map(error) error
}
