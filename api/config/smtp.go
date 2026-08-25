package config

import (
	"fmt"
)

// SMTPConfig holds email/SMTP server configuration
type SMTPConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	FromName  string `mapstructure:"from_name"`
	FromEmail string `mapstructure:"from_email"`
}

// GetSMTPAddress returns host:port formatted address
func (c *SMTPConfig) GetSMTPAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsConfigured checks if SMTP is properly configured
func (c *SMTPConfig) IsConfigured() bool {
	return c.Host != "" && c.Port > 0 && c.Username != "" && c.Password != ""
}
