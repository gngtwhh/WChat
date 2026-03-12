package config

import (
    "os"
    "sync"
    "time"
    "wchat/pkg/zlog"

    "github.com/BurntSushi/toml"
    "go.uber.org/zap"
)

type ServerConfig struct {
    AppName       string        `toml:"appName"`
    Host          string        `toml:"host"`
    Port          int           `toml:"port"`
    JwtSecret     string        `toml:"jwtSecret"`
    JwtExpireTime time.Duration `toml:"jwtExpireTime"`
    // SensitiveWordsFile string `toml:"sensitiveWordsFile"`
}

type MysqlConfig struct {
    Host         string `toml:"host"`
    Port         int    `toml:"port"`
    User         string `toml:"user"`
    Password     string `toml:"password"`
    DatabaseName string `toml:"databaseName"`
}

type RedisConfig struct {
    Host     string `toml:"host"`
    Port     int    `toml:"port"`
    Password string `toml:"password"`
    Db       int    `toml:"db"`
}

type AuthCodeConfig struct {
    AccessKeyID     string `toml:"accessKeyID"`
    AccessKeySecret string `toml:"accessKeySecret"`
    SignName        string `toml:"signName"`
    TemplateCode    string `toml:"templateCode"`
}

type LogConfig struct {
    LogPath string `toml:"logPath"`
}

type KafkaConfig struct {
    MessageMode string        `toml:"messageMode"`
    HostPort    string        `toml:"hostPort"`
    LoginTopic  string        `toml:"loginTopic"`
    LogoutTopic string        `toml:"logoutTopic"`
    ChatTopic   string        `toml:"chatTopic"`
    Partition   int           `toml:"partition"`
    Timeout     time.Duration `toml:"timeout"`
}

type StaticSrcConfig struct {
    StaticAvatarPath string `toml:"staticAvatarPath"`
    StaticFilePath   string `toml:"staticFilePath"`
}

type Config struct {
    ServerConfig    `toml:"mainConfig"`
    MysqlConfig     `toml:"mysqlConfig"`
    RedisConfig     `toml:"redisConfig"`
    AuthCodeConfig  `toml:"authCodeConfig"`
    LogConfig       `toml:"logConfig"`
    KafkaConfig     `toml:"kafkaConfig"`
    StaticSrcConfig `toml:"staticSrcConfig"`
}

var (
    config *Config
    once   sync.Once
)

func GetConfig() *Config {
    once.Do(
        func() {
            config = new(Config)
            // try env first
            filePath := os.Getenv("CONFIG_PATH")
            if filePath == "" {
                filePath = "configs/config.toml"
            }

            // 云服务器部署
            // if _, err := toml.DecodeFile("/root/project/KamaChat/configs/config_local.toml", config); err != nil {
            //     log.Fatal(err.Error())
            //     return err
            // }
            // 本地部署
            if _, err := toml.DecodeFile(filePath, config); err != nil {
                zlog.Fatal("Config load failed", zap.String("file", filePath), zap.Error(err))
            }
            zlog.Info("Config loaded successfully", zap.String("file", filePath))
        },
    )
    return config
}
