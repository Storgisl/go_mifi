package config

import (
    "os"
    "strconv"
)

type Config struct {
    DBHost        string
    DBPort        int
    DBUser        string
    DBPassword    string
    DBName        string
    JWTSecret     string
    SMTPHost      string
    SMTPPort      int
    SMTPUser      string
    SMTPPass      string
    PGPPublicKey  string
    PGPPrivateKey string
    HMACSecret    string
}

func LoadConfig() *Config {
    return &Config{
        DBHost:        getEnv("DB_HOST", "localhost"),
        DBPort:        getEnvInt("DB_PORT", 5432),
        DBUser:        getEnv("DB_USER", "bankuser"),
        DBPassword:    getEnv("DB_PASSWORD", "bankpass"),
        DBName:        getEnv("DB_NAME", "bankdb"),
        JWTSecret:     getEnv("JWT_SECRET", "supersecret"),
        SMTPHost:      getEnv("SMTP_HOST", "smtp.example.com"),
        SMTPPort:      getEnvInt("SMTP_PORT", 587),
        SMTPUser:      getEnv("SMTP_USER", "noreply@example.com"),
        SMTPPass:      getEnv("SMTP_PASS", "secret"),
        PGPPublicKey:  getEnv("PGP_PUBLIC_KEY", ""),
        PGPPrivateKey: getEnv("PGP_PRIVATE_KEY", ""),
        HMACSecret:    getEnv("HMAC_SECRET", "hmacsecret"),
    }
}

func getEnv(key, def string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return def
}

func getEnvInt(key string, def int) int {
    if val := os.Getenv(key); val != "" {
        if i, err := strconv.Atoi(val); err == nil {
            return i
        }
    }
    return def
}
