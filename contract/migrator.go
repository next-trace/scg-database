package contract

type (
	Migrator interface {
		Up() error
		Down(int) error
		Fresh() error
		Close() (sourceErr, dbErr error)
	}
)
