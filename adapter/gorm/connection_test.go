package gorm

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/next-trace/scg-database/config"
	"github.com/next-trace/scg-database/contract"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type (
	// dummyModel is a minimal model for connection tests.
	dummyModel struct{ contract.BaseModel }
)

func (d *dummyModel) TableName() string { return "dummy_models" }

// setupConnTestDB is a helper to create a GORM DB with a mock for connection-level tests.
func setupConnTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	// Removed VERSION query expectation since it's not consumed by most tests

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DisableAutomaticPing: true, // important for mock tests
	})
	require.NoError(t, err)
	return gormDB, mock
}

func TestConnection_Ping(t *testing.T) {
	gormDB, mock := setupConnTestDB(t)
	conn := &connection{db: gormDB}

	// --- Success case ---
	mock.ExpectPing()
	require.NoError(t, conn.Ping(t.Context()), "Ping should succeed")
	require.NoError(t, mock.ExpectationsWereMet())

	// --- Error case ---
	pingErr := errors.New("ping failed")
	mock.ExpectPing().WillReturnError(pingErr)
	err := conn.Ping(t.Context())
	require.Error(t, err)
	require.ErrorIs(t, err, pingErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConnection_Close(t *testing.T) {
	gormDB, mock := setupConnTestDB(t)
	conn := &connection{db: gormDB}

	mock.ExpectClose()
	require.NoError(t, conn.Close())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConnection_NewRepository(t *testing.T) {
	gormDB, _ := setupConnTestDB(t)
	conn := &connection{db: gormDB}

	repo, err := conn.NewRepository(&dummyModel{})
	require.NoError(t, err)
	require.NotNil(t, repo)

	_, err = conn.NewRepository(nil)
	require.Error(t, err, "NewRepository with nil model should fail")
	require.Equal(t, "model cannot be nil", err.Error())
}

func TestConnection_Transaction(t *testing.T) {
	t.Run("Commit", func(t *testing.T) {
		gormDB, mock := setupConnTestDB(t)
		conn := &connection{db: gormDB}
		mock.ExpectBegin()
		mock.ExpectCommit()

		err := conn.Transaction(t.Context(), func(tx contract.Connection) error {
			require.NotNil(t, tx, "Transaction connection should not be nil")
			return nil // Success
		})
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Rollback", func(t *testing.T) {
		gormDB, mock := setupConnTestDB(t)
		conn := &connection{db: gormDB}
		txErr := errors.New("transaction error")

		mock.ExpectBegin()
		// No commit, expect a rollback
		mock.ExpectRollback()

		err := conn.Transaction(t.Context(), func(_ contract.Connection) error {
			return txErr // Return an error to trigger rollback
		})
		require.Error(t, err)
		require.ErrorIs(t, err, txErr)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestConnection_RawQueries(t *testing.T) {
	gormDB, mock := setupConnTestDB(t)
	conn := &connection{db: gormDB, config: config.Config{}}
	ctx := t.Context()

	// --- Test Select ---
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT id FROM dummy_models").WillReturnRows(rows)
	results, err := conn.Select(ctx, "SELECT id FROM dummy_models")
	require.NoError(t, err)
	require.Len(t, results, 1)

	// --- Test Statement ---
	mock.ExpectExec("UPDATE dummy_models").WillReturnResult(sqlmock.NewResult(0, 1))
	res, err := conn.Statement(ctx, "UPDATE dummy_models SET name = 'foo'")
	require.NoError(t, err)
	affected, _ := res.RowsAffected()
	require.Equal(t, int64(1), affected)
	lastID, err := res.LastInsertId()
	require.NoError(t, err, "LastInsertId should work with sqlmock")
	require.Equal(t, int64(0), lastID)

	// --- Test Error Path ---
	dbErr := errors.New("db error")
	mock.ExpectQuery("SELECT id FROM dummy_models").WillReturnError(dbErr)
	_, err = conn.Select(ctx, "SELECT id FROM dummy_models")
	require.Error(t, err)
	require.ErrorIs(t, err, dbErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConnection_GetConnection(t *testing.T) {
	gormDB, _ := setupConnTestDB(t)
	conn := &connection{db: gormDB}
	rawDB := conn.GetConnection()
	require.NotNil(t, rawDB)
	_, ok := rawDB.(*gorm.DB)
	require.True(t, ok, "GetConnection should return the underlying *gorm.DB instance")
}
