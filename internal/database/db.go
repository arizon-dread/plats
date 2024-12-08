package database

type Db interface {
	Store(string, any) error
	Get(string) *string
}
