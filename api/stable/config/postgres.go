package config

import "time"

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// GetMaxOpenConns returns max open connections with default
func (p *PostgresConfig) GetMaxOpenConns() int {
	if p.MaxOpenConns == 0 {
		return 25
	}
	return p.MaxOpenConns
}

// GetMaxIdleConns returns max idle connections with default
func (p *PostgresConfig) GetMaxIdleConns() int {
	if p.MaxIdleConns == 0 {
		return 5
	}
	return p.MaxIdleConns
}

// GetConnMaxLifetime returns connection max lifetime with default
func (p *PostgresConfig) GetConnMaxLifetime() time.Duration {
	if p.ConnMaxLifetime == 0 {
		return 5 * time.Minute
	}
	return p.ConnMaxLifetime
}
