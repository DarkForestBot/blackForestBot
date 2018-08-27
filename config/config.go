package config

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"path/filepath"
)

//Config for this bot
type Config struct {
	APIToken      string `json:"tgApiToken"`
	Debug         bool   `json:"debug"`
	UpdateTimeout int    `json:"updateTimeout"`
	Database      string `json:"database"`
	Redis         string `json:"redis"`
	Images        *struct {
		Win     string `json:"win"`
		Lose    string `json:"lose"`
		Start   string `json:"start"`
		Killed  string `json:"killed"`
		Trapped string `json:"trapped"`
	} `json:"img"`
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

func init() {
	appPath, _ := filepath.Abs(path.Dir(os.Args[0]))
	log.Printf("Find config in `%s`\n", appPath)
	if err := DefaultConfig.ReadConfig(path.Join(appPath, "config.json")); err != nil {
		log.Panicln(err)
	}
	log.Printf("Config loaded.")
}
