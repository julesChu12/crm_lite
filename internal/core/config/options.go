package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// GinMode Gin框架运行模式
type GinMode string

const (
	DebugMode   GinMode = "debug"   // 调试模式，输出详细日志
	ReleaseMode GinMode = "release" // 发布模式，性能优化
	TestMode    GinMode = "test"    // 测试模式
)

// ==================== 配置结构体定义 ====================

// ServerTLSConfig HTTPS/WSS 证书配置
type ServerTLSConfig struct {
	CertFile string `mapstructure:"certFile"` // SSL证书文件路径
	KeyFile  string `mapstructure:"keyFile"`  // SSL私钥文件路径
}

// ServerOptions 服务器相关配置
type ServerOptions struct {
	Mode    GinMode         `mapstructure:"mode"`    // 运行模式
	Host    string          `mapstructure:"host"`    // 监听地址
	Port    uint16          `mapstructure:"port"`    // 监听端口
	Timeout time.Duration   `mapstructure:"timeout"` // HTTP超时时间
	WSAddr  string          `mapstructure:"wsAddr"`  // WebSocket地址
	WSSAddr string          `mapstructure:"wssAddr"` // 安全WebSocket地址
	TLS     ServerTLSConfig `mapstructure:"tls"`     // TLS配置
	PidFile string          `mapstructure:"pidFile"` // PID文件路径
}

// LogOptions 日志配置
type LogOptions struct {
	Dir                string `mapstructure:"dir"`                // 日志目录
	Level              int8   `mapstructure:"level"`              // 日志级别
	LineNum            bool   `mapstructure:"lineNum"`            // 是否显示行号
	Filename           string `mapstructure:"filename"`           // 日志文件名
	MaxSize            int    `mapstructure:"maxSize"`            // 单文件最大大小(字节)
	MaxAge             int    `mapstructure:"maxAge"`             // 保存天数
	MaxBackups         int    `mapstructure:"maxBackups"`         // 最大备份数
	Compress           bool   `mapstructure:"compress"`           // 是否压缩
	EnableTimeRotation bool   `mapstructure:"enableTimeRotation"` // 启用按时间轮转
	RotationTime       string `mapstructure:"rotationTime"`       // 轮转间隔: hourly, daily, weekly
	LinkName           string `mapstructure:"linkName"`           // 当前日志文件的软链接名称
}

// LogCleanupOptions 日志清理配置
type LogCleanupOptions struct {
	Mode         string        `mapstructure:"mode"`         // 清理模式: internal, external
	Interval     time.Duration `mapstructure:"interval"`     // 检查间隔
	RetentionDay int           `mapstructure:"retentionDay"` // 日志保留天数
	DryRun       bool          `mapstructure:"dryRun"`       // 试运行模式
}

// DBOptions 数据库配置
type DBOptions struct {
	Driver          string        `mapstructure:"driver"`          // 数据库驱动
	Host            string        `mapstructure:"host"`            // 主机地址
	Port            int           `mapstructure:"port"`            // 端口
	User            string        `mapstructure:"user"`            // 用户名
	Password        string        `mapstructure:"password"`        // 密码
	DBName          string        `mapstructure:"dbname"`          // 数据库名
	TablePrefix     string        `mapstructure:"tablePrefix"`     // 表前缀
	SSLMode         string        `mapstructure:"sslmode"`         // SSL模式
	TimeZone        string        `mapstructure:"timeZone"`        // 时区
	DSN             string        `mapstructure:"dsn"`             // 兼容测试场景直接指定 DSN
	MaxOpenConns    int64         `mapstructure:"maxOpenConns"`    // 最大连接数
	MaxIdleConns    int64         `mapstructure:"maxIdleConns"`    // 最大空闲连接数
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"` // 连接最大存活时间
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `mapstructure:"host"`         // 主机地址
	Port         int           `mapstructure:"port"`         // 端口
	Password     string        `mapstructure:"password"`     // 密码
	DB           int           `mapstructure:"db"`           // 数据库编号
	MaxRetries   int           `mapstructure:"maxRetries"`   // 最大重试次数
	PoolSize     int           `mapstructure:"poolSize"`     // 连接池大小
	MinIdleConns int           `mapstructure:"minIdleConns"` // 最小空闲连接数
	MaxConnAge   time.Duration `mapstructure:"maxConnAge"`   // 连接最大存活时间
	IdleTimeout  time.Duration `mapstructure:"idleTimeout"`  // 空闲超时时间
	DialTimeout  time.Duration `mapstructure:"dialTimeout"`  // 连接超时时间
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`  // 读取超时时间
	WriteTimeout time.Duration `mapstructure:"writeTimeout"` // 写入超时时间
}

// CacheOptions 缓存配置
type CacheOptions struct {
	Driver string      `mapstructure:"driver"` // 缓存驱动类型
	Redis  RedisConfig `mapstructure:"redis"`  // Redis配置
}

// JWTOptions JWT认证配置
type JWTOptions struct {
	Secret             string        `mapstructure:"secret"`               // 签名密钥
	Issuer             string        `mapstructure:"issuer"`               // 发行者
	Algorithm          string        `mapstructure:"algorithm"`            // 签名算法
	ExpireDuration     time.Duration `mapstructure:"expire_duration"`      // 默认过期时间
	AccessTokenExpire  time.Duration `mapstructure:"access_token_expire"`  // 访问令牌过期时间
	RefreshTokenExpire time.Duration `mapstructure:"refresh_token_expire"` // 刷新令牌过期时间
}

// RbacOptions RBAC权限配置
type RbacOptions struct {
	ModelFile string `mapstructure:"modelFile"` // Casbin模型文件路径
}

// SuperAdminOptions 超级管理员账号配置
type SuperAdminOptions struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
	Role     string `mapstructure:"role"`
}

// EmailOptions 邮件服务配置
type EmailOptions struct {
	Host         string `mapstructure:"host"`     // SMTP 服务器地址
	Port         int    `mapstructure:"port"`     // SMTP 服务器端口
	Username     string `mapstructure:"username"` // 邮箱账号
	Password     string `mapstructure:"password"` // 邮箱密码或授权码
	FromAddress  string `mapstructure:"from"`     // 发件人邮箱地址
	FromName     string `mapstructure:"fromName"` // 发件人名称
	UseTLS       bool   `mapstructure:"useTLS"`   // 是否使用TLS加密
	InsecureSkip bool   `mapstructure:"insecureSkip"`
}

// AuthOptions 认证授权配置
type AuthOptions struct {
	JWTOptions   `mapstructure:"jwt"`  // JWT配置
	RbacOptions  `mapstructure:"rbac"` // RBAC配置
	OAuthEnabled bool                  `mapstructure:"oauthEnabled"` // 是否启用OAuth
	BCryptCost   int                   `mapstructure:"BCryptCost"`   // 密码加密成本
	SuperAdmin   SuperAdminOptions     `mapstructure:"superAdmin"`   // 超级管理员配置
	Email        EmailOptions          `mapstructure:"email"`        // 邮件服务配置
	DefaultRole  string                `mapstructure:"defaultRole"`  // 默认角色名称
}

// ==================== 主配置结构体 ====================

// Options 应用程序配置选项
type Options struct {
	vp *viper.Viper // viper实例

	Server     ServerOptions     `mapstructure:"server"`     // 服务器配置
	Logger     LogOptions        `mapstructure:"logger"`     // 日志配置
	LogCleanup LogCleanupOptions `mapstructure:"logCleanup"` // 日志清理配置
	Database   DBOptions         `mapstructure:"database"`   // 数据库配置
	Cache      CacheOptions      `mapstructure:"cache"`      // 缓存配置
	Auth       AuthOptions       `mapstructure:"auth"`       // 认证配置
	Email      EmailOptions      `mapstructure:"email"`      // 邮件服务配置
	PprofOn    bool              `mapstructure:"pprofOn"`    // 性能分析开关
}

// ==================== 单例模式 ====================

var (
	instance *Options
	once     sync.Once
)

// GetInstance 获取配置单例实例
func GetInstance() *Options {
	once.Do(func() {
		instance = &Options{}
	})
	return instance
}

// SetInstanceForTest 仅用于测试，允许重置配置实例
func SetInstanceForTest(opts *Options) {
	instance = opts
	// 重置 once，以便 GetInstance 可以在下次调用时重新初始化
	once = sync.Once{}
}

// ==================== 配置初始化 ====================

// InitOptions 初始化配置选项
func InitOptions(configFile string) error {
	opts := GetInstance()

	// 创建viper实例
	vp := viper.New()
	vp.SetConfigFile(configFile)

	// 读取配置文件
	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %v", err)
		}
	}

	// 启用环境变量自动映射
	vp.AutomaticEnv()
	// 开启环境变量自动映射后 优先级则从高到低为 Set() > 环境变量 > flag > 配置文件 > 默认值 ： db config 获取的是.env 中的值
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	// 应用配置
	opts.ConfigureWithViper(vp)
	return nil
}

// ConfigureWithViper 使用viper配置系统选项
func (o *Options) ConfigureWithViper(vp *viper.Viper) {
	o.vp = vp

	// 服务器配置
	o.Server = ServerOptions{
		Mode:    GinMode(o.getStringWithDefault("server.mode", string(DebugMode))),
		Host:    o.getStringWithDefault("server.host", "0.0.0.0"),
		Port:    o.getUInt16WithDefault("server.port", 8080),
		Timeout: o.getDurationWithDefault("server.timeout", 30*time.Second),
		WSAddr:  o.getStringWithDefault("server.wsAddr", ""),
		WSSAddr: o.getStringWithDefault("server.wssAddr", ""),
		TLS: ServerTLSConfig{
			CertFile: o.getStringWithDefault("server.tls.certFile", ""),
			KeyFile:  o.getStringWithDefault("server.tls.keyFile", ""),
		},
		PidFile: o.getStringWithDefault("server.pidFile", "server.pid"),
	}

	// 日志配置
	o.Logger = LogOptions{
		Dir:        o.getStringWithDefault("logger.dir", ""),
		Level:      int8(o.getIntWithDefault("logger.level", 0)),
		Filename:   o.getStringWithDefault("logger.filename", "app.log"),
		LineNum:    o.getBoolWithDefault("logger.lineNum", false),
		MaxSize:    o.getIntWithDefault("logger.maxSize", 100*1024*1024),
		MaxAge:     o.getIntWithDefault("logger.maxAge", 7),
		MaxBackups: o.getIntWithDefault("logger.maxBackups", 3),
		Compress:   o.getBoolWithDefault("logger.compress", false),
	}

	// 数据库配置
	o.Database = DBOptions{
		Driver:          o.getStringWithDefault("db.driver", "mysql"),
		Host:            o.getStringWithDefault("db.host", "127.0.0.1"),
		Port:            o.getIntWithDefault("db.port", 3306),
		User:            o.getStringWithDefault("db.user", "root"),
		Password:        o.getStringWithDefault("db.password", ""),
		DBName:          o.getStringWithDefault("db.dbname", "test_db"),
		TablePrefix:     o.getStringWithDefault("db.tablePrefix", ""),
		SSLMode:         o.getStringWithDefault("db.sslmode", "disable"),
		TimeZone:        o.getStringWithDefault("db.timeZone", "Asia/Shanghai"),
		DSN:             o.getStringWithDefault("db.dsn", ""), // 兼容测试场景直接指定 DSN
		MaxOpenConns:    o.getInt64WithDefault("db.maxOpenConns", 100),
		MaxIdleConns:    o.getInt64WithDefault("db.maxIdleConns", 10),
		ConnMaxLifetime: o.getDurationWithDefault("db.connMaxLifetime", time.Hour),
	}

	// 缓存配置
	o.Cache = CacheOptions{
		Driver: o.getStringWithDefault("cache.driver", "redis"),
		Redis: RedisConfig{
			Host:         o.getStringWithDefault("cache.redis.host", "127.0.0.1"),
			Port:         o.getIntWithDefault("cache.redis.port", 6379),
			Password:     o.getStringWithDefault("cache.redis.password", ""),
			DB:           o.getIntWithDefault("cache.redis.db", 0),
			MaxRetries:   o.getIntWithDefault("cache.redis.maxRetries", 3),
			PoolSize:     o.getIntWithDefault("cache.redis.poolSize", 10),
			MinIdleConns: o.getIntWithDefault("cache.redis.minIdleConns", 5),
			MaxConnAge:   o.getDurationWithDefault("cache.redis.maxConnAge", time.Hour),
			IdleTimeout:  o.getDurationWithDefault("cache.redis.idleTimeout", 5*time.Minute),
			DialTimeout:  o.getDurationWithDefault("cache.redis.dialTimeout", 5*time.Second),
			ReadTimeout:  o.getDurationWithDefault("cache.redis.readTimeout", 3*time.Second),
			WriteTimeout: o.getDurationWithDefault("cache.redis.writeTimeout", 3*time.Second),
		},
	}

	// 认证配置
	o.Auth = AuthOptions{
		JWTOptions: JWTOptions{
			Secret:             o.getStringWithDefault("auth.jwt.secret", "default-secret"),
			Issuer:             o.getStringWithDefault("auth.jwt.issuer", "crm-lite"),
			Algorithm:          o.getStringWithDefault("auth.jwt.algorithm", "HS256"),
			ExpireDuration:     o.getDurationWithDefault("auth.jwt.expire_duration", 3600*time.Second),
			AccessTokenExpire:  o.getDurationWithDefault("auth.jwt.access_token_expire", 3600*time.Second),
			RefreshTokenExpire: o.getDurationWithDefault("auth.jwt.refresh_token_expire", 3600*time.Second),
		},
		RbacOptions: RbacOptions{
			ModelFile: o.getStringWithDefault("auth.rbac.modelFile", "config/rbac_model.conf"),
		},
		OAuthEnabled: o.getBoolWithDefault("auth.oauthEnabled", false),
		BCryptCost:   o.getIntWithDefault("auth.BCryptCost", 10),
		SuperAdmin: SuperAdminOptions{
			Username: o.getStringWithDefault("auth.superAdmin.username", "admin"),
			Password: o.getStringWithDefault("auth.superAdmin.password", "admin"),
			Email:    o.getStringWithDefault("auth.superAdmin.email", "admin@example.com"),
			Role:     o.getStringWithDefault("auth.superAdmin.role", "super_admin"),
		},
		Email: EmailOptions{
			Host:         o.getStringWithDefault("email.host", ""),
			Port:         o.getIntWithDefault("email.port", 587),
			Username:     o.getStringWithDefault("email.username", ""),
			Password:     o.getStringWithDefault("email.password", ""),
			FromAddress:  o.getStringWithDefault("email.from", ""),
			FromName:     o.getStringWithDefault("email.fromName", "CRM Lite"),
			UseTLS:       o.getBoolWithDefault("email.useTLS", true),
			InsecureSkip: o.getBoolWithDefault("email.insecureSkip", true),
		},
		DefaultRole: o.getStringWithDefault("auth.defaultRole", ""),
	}
	// 其他配置
	o.PprofOn = o.getBoolWithDefault("pprofOn", false)
}

// ==================== 辅助方法 ====================

// getStringWithDefault 获取字符串配置值，提供默认值
func (o *Options) getStringWithDefault(key, defaultValue string) string {
	if o.vp.IsSet(key) {
		return o.vp.GetString(key)
	}
	return defaultValue
}

// getIntWithDefault 获取整数配置值，提供默认值
func (o *Options) getIntWithDefault(key string, defaultValue int) int {
	if o.vp.IsSet(key) {
		return o.vp.GetInt(key)
	}
	return defaultValue
}

// getInt64WithDefault 获取int64配置值，提供默认值
func (o *Options) getInt64WithDefault(key string, defaultValue int64) int64 {
	if o.vp.IsSet(key) {
		return o.vp.GetInt64(key)
	}
	return defaultValue
}

// getUInt16WithDefault 获取uint16配置值，提供默认值
func (o *Options) getUInt16WithDefault(key string, defaultValue uint16) uint16 {
	if o.vp.IsSet(key) {
		return cast.ToUint16(o.vp.Get(key))
	}
	return defaultValue
}

// getDurationWithDefault 获取时间配置值，提供默认值
func (o *Options) getDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	if o.vp.IsSet(key) {
		return o.vp.GetDuration(key)
	}
	return defaultValue
}

// getBoolWithDefault 获取布尔配置值，提供默认值
func (o *Options) getBoolWithDefault(key string, defaultValue bool) bool {
	if o.vp.IsSet(key) {
		return o.vp.GetBool(key)
	}
	return defaultValue
}
