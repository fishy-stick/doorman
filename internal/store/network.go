package store

import (
	"time"

	"gorm.io/gorm"
)

type Network struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Token       string    `json:"token" gorm:"not null"`
	DDNSEnabled bool      `json:"ddns_enabled" gorm:"default:false"`
	DDNSType    string    `json:"ddns_type" gorm:"default:''"`
	DDNSConfig  string    `json:"ddns_config" gorm:"default:'{}'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NetworkSummary struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	DDNSEnabled bool       `json:"ddns_enabled"`
	DDNSType    string     `json:"ddns_type"`
	CurrentIP   *string    `json:"current_ip"`
	PreviousIP  *string    `json:"previous_ip"`
	LastKnock   *time.Time `json:"last_knock"`
	DDNSStatus  *string    `json:"ddns_status"`
}

func (s *Store) ListNetworks() ([]NetworkSummary, error) {
	var networks []NetworkSummary

	subQuery := s.db.Model(&Knock{}).
		Select("network_id, ip, previous_ip, created_at, ddns_status, ROW_NUMBER() OVER (PARTITION BY network_id ORDER BY created_at DESC) as rn").
		Table("knocks")

	err := s.db.Model(&Network{}).
		Select("networks.id, networks.name, networks.ddns_enabled, networks.ddns_type, c.ip as current_ip, c.previous_ip, c.created_at as last_knock, c.ddns_status").
		Joins("LEFT JOIN (?) c ON c.network_id = networks.id AND c.rn = 1", subQuery).
		Order("networks.name").
		Scan(&networks).Error

	return networks, err
}

func (s *Store) GetNetwork(id int64) (*Network, error) {
	var network Network
	if err := s.db.First(&network, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &network, nil
}

func (s *Store) GetNetworkByToken(token string) (*Network, error) {
	var network Network
	if err := s.db.Where("token = ?", token).First(&network).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &network, nil
}

func (s *Store) CreateNetwork(n *Network) error {
	return s.db.Create(n).Error
}

func (s *Store) UpdateNetwork(n *Network) error {
	return s.db.Save(n).Error
}

func (s *Store) DeleteNetwork(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("network_id = ?", id).Delete(&Knock{}).Error; err != nil {
			return err
		}
		return tx.Delete(&Network{}, id).Error
	})
}
