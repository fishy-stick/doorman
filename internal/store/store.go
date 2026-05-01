package store

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

// Store wraps the application's persistent state.
type Store struct {
	db *gorm.DB
}

// New opens the SQLite database and runs schema migrations.
func New(dbPath string) (*Store, error) {
	// Use the pure-Go modernc SQLite driver so the service and tests do not
	// depend on a local CGO toolchain.
	db, err := gorm.Open(sqlite.New(sqlite.Config{
		DriverName: "sqlite",
		DSN:        dbPath,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Network{}, &Knock{}, &Admin{}); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
