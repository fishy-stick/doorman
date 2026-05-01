package store

import (
	"time"

	"gorm.io/gorm"
)

type Knock struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	NetworkID  int64     `json:"network_id" gorm:"index;not null"`
	IP         string    `json:"ip" gorm:"not null"`
	PreviousIP *string   `json:"previous_ip"`
	IPChanged  bool      `json:"ip_changed" gorm:"default:false"`
	UserAgent  string    `json:"user_agent" gorm:"default:''"`
	DDNSStatus string    `json:"ddns_status" gorm:"default:'skipped'"`
	DDNSError  string    `json:"ddns_error" gorm:"default:''"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *Store) InsertKnock(c *Knock) error {
	return s.db.Create(c).Error
}

func (s *Store) GetLatestKnock(networkID int64) (*Knock, error) {
	var knock Knock
	if err := s.db.Where("network_id = ?", networkID).Order("created_at DESC").First(&knock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &knock, nil
}

func (s *Store) GetPreviousIP(networkID int64) (*string, error) {
	var knock Knock
	if err := s.db.Where("network_id = ?", networkID).Order("created_at DESC").First(&knock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	if knock.PreviousIP != nil {
		return knock.PreviousIP, nil
	}
	return nil, nil
}

func (s *Store) ListKnocks(networkID int64, page, size int) ([]Knock, int, error) {
	var total int64
	if err := s.db.Model(&Knock{}).Where("network_id = ?", networkID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var knocks []Knock
	offset := (page - 1) * size
	if err := s.db.Where("network_id = ?", networkID).
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&knocks).Error; err != nil {
		return nil, 0, err
	}

	return knocks, int(total), nil
}
