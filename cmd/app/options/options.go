package options

/*
*
应用全局配置
*/
type ApplicationOptions struct {
	Redis *RedisOptions `mapstructure:"redis"`
	Log   *LogOptions   `mapstructure:"log"`
}

type RedisOptions struct {
	Addr     string `mapstructure:"addr"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	UserName string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type LogOptions struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Path   string `mapstructure:"path"`
}

func NewApplicationOptions() *ApplicationOptions {
	return &ApplicationOptions{
		Redis: &RedisOptions{},
		Log:   &LogOptions{},
	}
}
