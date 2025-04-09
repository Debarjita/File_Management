package jobs

import (
	"log"
	"os"
	"time"

	"learningfilesharing/config"
	"learningfilesharing/models"
)

func StartCleanupJob() {
	ticker := time.NewTicker(30 * time.Minute) // adjust as needed

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("🔁 Running cleanup job...")

				var expiredFiles []models.File
				if err := config.DB.Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).Find(&expiredFiles).Error; err != nil {
					log.Println(" Failed to fetch expired files:", err)
					continue
				}

				for _, file := range expiredFiles {
					// Delete from local storage
					err := os.Remove("." + file.URL) // assuming /uploads/file.ext
					if err != nil {
						log.Println(" Failed to delete file:", file.URL, err)
					}

					// Delete from DB
					if err := config.DB.Delete(&file).Error; err != nil {
						log.Println(" Failed to delete metadata:", file.ID, err)
					} else {
						log.Println("Deleted:", file.FileName)
					}
				}
			}
		}
	}()
}
