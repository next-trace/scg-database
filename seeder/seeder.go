package seeder

import (
	"fmt"

	"github.com/next-trace/scg-database/contract"
)

type (
	Runner struct{ connection contract.Connection }
)

func New(db contract.Connection) *Runner { return &Runner{connection: db} }
func (r *Runner) Run(seeders ...contract.Seeder) error {
	for _, s := range seeders {
		if err := s.Run(r.connection); err != nil {
			return fmt.Errorf("failed to run seeder %T: %w", s, err)
		}
	}
	return nil
}
