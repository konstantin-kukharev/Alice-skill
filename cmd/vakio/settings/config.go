package settings

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	DefaultServerAddr     = "0.0.0.0:8080"
	DefaultPoolInterval   = 300 // время между запросами в секундах
	DefaultReportInterval = 600 // время между запросами в секундах
)

type VakioConfig struct {
	PoolInterval   time.Duration // время между запросами
	ReportInterval time.Duration // время между запросами
	Login          string        // login
	CID            string        // client id
	Secret         string        // client secret
	Password       string        // password
}

type Config struct {
	Address string // адрес эндпоинта HTTP-сервера
	VakioConfig
}

// Если указана переменная окружения, то используется она.
// Если нет переменной окружения, но есть аргумент командной строки (флаг), то используется он.
// Если нет ни переменной окружения, ни флага, то используется значение по умолчанию.
func New() *Config {
	c := &Config{
		Address: DefaultServerAddr,
		VakioConfig: VakioConfig{
			PoolInterval:   DefaultPoolInterval,
			ReportInterval: DefaultReportInterval,
			Login:          "",
			CID:            "",
			Secret:         "",
			Password:       "",
		},
	}

	return c
}

// ADDRESS отвечает за адрес эндпоинта HTTP-сервера.
// VAKIO_REPORT_INTERVAL позволяет переопределять reportInterval.
// VAKIO_POLL_INTERVAL позволяет переопределять pollInterval.
// VAKIO_CID позволяет переопределять pollInterval.
// VAKIO_LOGIN позволяет переопределять pollInterval.
// VAKIO_SECRET позволяет переопределять pollInterval.
// VAKIO_PASSWORD позволяет переопределять pollInterval.
func (c *Config) WithEnv() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки файла .env:", err)
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		c.Address = envRunAddr
	}

	if envReportInterval := os.Getenv("VAKIO_REPORT_INTERVAL"); envReportInterval != "" {
		if val, err := strconv.Atoi(envReportInterval); err == nil {
			c.ReportInterval = time.Duration(val) * time.Second
		}
	}

	if envPoolInterval := os.Getenv("VAKIO_POLL_INTERVAL"); envPoolInterval != "" {
		if val, err := strconv.Atoi(envPoolInterval); err == nil {
			c.PoolInterval = time.Duration(val) * time.Second
		}
	}

	if envVakioCID := os.Getenv("VAKIO_CID"); envVakioCID != "" {
		c.CID = envVakioCID
	}

	if envVakioLogin := os.Getenv("VAKIO_LOGIN"); envVakioLogin != "" {
		c.Login = envVakioLogin
	}

	if envVakioSecret := os.Getenv("VAKIO_SECRET"); envVakioSecret != "" {
		c.Secret = envVakioSecret
	}

	if envVakioPassword := os.Getenv("VAKIO_PASSWORD"); envVakioPassword != "" {
		c.Password = envVakioPassword
	}

	return c
}
