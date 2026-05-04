package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidNetwork       = errors.New("invalid network")
	ErrNetworkNameConflict  = errors.New("network name already exists")
	ErrNetworkTokenConflict = errors.New("network token already exists")
)

// Network represents a managed internal network and its DDNS settings.
type Network struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Token       string    `json:"token" gorm:"uniqueIndex;not null"`
	DDNSEnabled bool      `json:"ddns_enabled" gorm:"column:ddns_enabled;default:false"`
	DDNSType    string    `json:"ddns_type" gorm:"column:ddns_type;default:''"`
	DDNSConfig  string    `json:"ddns_config" gorm:"column:ddns_config;default:'{}'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NetworkSummary is the admin list projection with the latest knock snapshot.
type NetworkSummary struct {
	ID          int64      `json:"id" gorm:"column:id"`
	Name        string     `json:"name" gorm:"column:name"`
	DDNSEnabled bool       `json:"ddns_enabled" gorm:"column:ddns_enabled"`
	DDNSType    string     `json:"ddns_type" gorm:"column:ddns_type"`
	CurrentIP   *string    `json:"current_ip" gorm:"column:current_ip"`
	PreviousIP  *string    `json:"previous_ip" gorm:"column:previous_ip"`
	LastKnock   *time.Time `json:"last_knock" gorm:"column:last_knock"`
	DDNSStatus  *string    `json:"ddns_status" gorm:"column:ddns_status"`
}

func (s *Store) ListNetworks() ([]NetworkSummary, error) {
	networks := make([]NetworkSummary, 0)

	subQuery := s.db.Model(&Knock{}).
		Select("network_id, ip, previous_ip, created_at, ddns_status, id, ROW_NUMBER() OVER (PARTITION BY network_id ORDER BY created_at DESC, id DESC) as rn").
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
	normalizeNetwork(n)
	if err := validateNetwork(n); err != nil {
		return err
	}
	if err := s.ensureNetworkUniqueFields(n); err != nil {
		return err
	}
	return s.db.Create(n).Error
}

func (s *Store) UpdateNetwork(n *Network) error {
	normalizeNetwork(n)
	if err := validateNetwork(n); err != nil {
		return err
	}
	if err := s.ensureNetworkUniqueFields(n); err != nil {
		return err
	}

	return s.db.Model(&Network{}).
		Where("id = ?", n.ID).
		Updates(map[string]any{
			"name":         n.Name,
			"token":        n.Token,
			"ddns_enabled": n.DDNSEnabled,
			"ddns_type":    n.DDNSType,
			"ddns_config":  n.DDNSConfig,
		}).Error
}

func (s *Store) UpdateNetworkToken(id int64, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("%w: token is required", ErrInvalidNetwork)
	}

	tokenExists, err := s.networkFieldExists("token", token, id)
	if err != nil {
		return err
	}
	if tokenExists {
		return ErrNetworkTokenConflict
	}

	return s.db.Model(&Network{}).
		Where("id = ?", id).
		Update("token", token).Error
}

func (s *Store) DeleteNetwork(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("network_id = ?", id).Delete(&Knock{}).Error; err != nil {
			return err
		}
		return tx.Delete(&Network{}, id).Error
	})
}

func normalizeNetwork(n *Network) {
	if n == nil {
		return
	}

	n.Name = strings.TrimSpace(n.Name)
	n.Token = strings.TrimSpace(n.Token)
	n.DDNSType = strings.TrimSpace(n.DDNSType)
	if !n.DDNSEnabled {
		n.DDNSType = ""
		n.DDNSConfig = "{}"
		return
	}
	if strings.TrimSpace(n.DDNSConfig) == "" {
		n.DDNSConfig = "{}"
	}
}

func validateNetwork(n *Network) error {
	if n == nil {
		return fmt.Errorf("%w: network is required", ErrInvalidNetwork)
	}
	if n.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidNetwork)
	}
	if n.Token == "" {
		return fmt.Errorf("%w: token is required", ErrInvalidNetwork)
	}
	if err := validateDDNSSettings(n.DDNSEnabled, n.DDNSType, n.DDNSConfig); err != nil {
		return err
	}

	return nil
}

func (s *Store) ensureNetworkUniqueFields(n *Network) error {
	nameExists, err := s.networkFieldExists("name", n.Name, n.ID)
	if err != nil {
		return err
	}
	if nameExists {
		return ErrNetworkNameConflict
	}

	tokenExists, err := s.networkFieldExists("token", n.Token, n.ID)
	if err != nil {
		return err
	}
	if tokenExists {
		return ErrNetworkTokenConflict
	}

	return nil
}

func (s *Store) networkFieldExists(field, value string, excludeID int64) (bool, error) {
	query := s.db.Model(&Network{}).Where(field+" = ?", value)
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
