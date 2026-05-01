package store

import (
	"time"

	"gorm.io/gorm"
)

// Knock stores one observed client request and the DDNS result for it.
type Knock struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	NetworkID int64  `json:"network_id" gorm:"index;not null"`
	IP        string `json:"ip" gorm:"not null"`
	// PreviousIP stores the last observed IP before this knock was recorded.
	PreviousIP *string   `json:"previous_ip"`
	IPChanged  bool      `json:"ip_changed" gorm:"default:false"`
	UserAgent  string    `json:"user_agent" gorm:"default:''"`
	DDNSStatus string    `json:"ddns_status" gorm:"column:ddns_status;default:'skipped'"`
	DDNSError  string    `json:"ddns_error" gorm:"column:ddns_error;default:''"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *Store) InsertKnock(c *Knock) error {
	return s.db.Create(c).Error
}

func (s *Store) GetLatestKnock(networkID int64) (*Knock, error) {
	var knock Knock
	if err := s.db.Where("network_id = ?", networkID).Order("created_at DESC").Order("id DESC").First(&knock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &knock, nil
}

// GetLatestIP returns the most recently observed IP for the network.
func (s *Store) GetLatestIP(networkID int64) (*string, error) {
	knock, err := s.GetLatestKnock(networkID)
	if err != nil {
		return nil, err
	}
	if knock == nil {
		return nil, nil
	}

	ip := knock.IP
	return &ip, nil
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
		Order("id DESC").
		Offset(offset).
		Limit(size).
		Find(&knocks).Error; err != nil {
		return nil, 0, err
	}

	return knocks, int(total), nil
}
