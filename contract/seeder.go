package contract

type (
	Seeder interface{ Run(Connection) error }
)
