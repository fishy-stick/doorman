package store

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

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

func New(dbPath string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Network{}, &Knock{}); err != nil {
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

// Network operations

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

// Knock operations

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
