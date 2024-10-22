package utils

import (
	"fmt"
	"service/all/internal/repository/config"
)

func GetDSN(db *config.DB) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		db.Host,
		db.User,
		db.Password,
		db.Name,
		db.Port,
	)
}
