package config

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/go-redis/redis"
)

//Config for this bot
type Config struct {
	APIToken      string `json:"tgApiToken"`
	Debug         bool   `json:"debug"`
	UpdateTimeout int    `json:"updateTimeout"`
	Database      string `json:"database"`
	Redis         string `json:"redis"`
	AdminPassword string `json:"adminPassword"`
}

//DefaultConfig is global
var DefaultConfig Config

//ReadConfig from file
func (c *Config) ReadConfig(cfgFile string) error {
	absCfgFile, err := filepath.Abs(cfgFile)
	if err != nil {
		return err
	}
	f, err := os.Open(absCfgFile)
	if err != nil {
		return err
	}
	log.Printf("Found config `config.json`\n")
	defer f.Close()
	j := json.NewDecoder(f)
	if err := j.Decode(c); err != nil {
		return err
	}
	return nil
}

// RedisConfig to ready
func (c *Config) RedisConfig() (*redis.Options, error) {
	u, err := url.Parse(c.Redis)
	if err != nil {
		return nil, err
	}
	o := new(redis.Options)
	o.Addr = u.Host
	_, db := path.Split(u.Path)
	dbn, err := strconv.Atoi(db)
	if err != nil {
		return nil, err
	}
	o.DB = dbn
	return o, nil
}

func init() {
	appPath, _ := filepath.Abs(path.Dir(os.Args[0]))
	log.Printf("Find config in `%s`\n", appPath)
	if err := DefaultConfig.ReadConfig(path.Join(appPath, "config.json")); err != nil {
		log.Panicln(err)
	}
	log.Printf("Config loaded.")
}
