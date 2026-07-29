package store

import (
	"github.com/alessandro/niedduty/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open verbindet zur Datenbank und migriert das Schema.
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Invite{},
		&models.Session{},
		&models.Club{},
		&models.Player{},
		&models.LeagueEntry{},
		&models.Penalty{},
		&models.PlayerPenalty{},
		&models.Event{},
		&models.EventAttendance{},
	); err != nil {
		return nil, err
	}
	return db, nil
}
