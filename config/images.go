package config

import (
	"encoding/gob"
	"log"
	"os"
	"path"
	"path/filepath"
)

// Images in game
type Images struct {
	Win     string
	Lose    string
	Start   string
	Killed  string
	Trapped string
}

var DefaultImages Images

func (i *Images) ReadConfig(cfgFile string) error {
	absCfgFile, err := filepath.Abs(cfgFile)
	if err != nil {
		return err
	}
	f, err := os.Open(absCfgFile)
	if err != nil {
		return err
	}
	log.Printf("Found config `images.dat`\n")
	defer f.Close()
	d := gob.NewDecoder(f)
	if err := d.Decode(i); err != nil {
		return err
	}
	return nil
}

func (i *Images) WriteConfig(cfgFile string) error {
	absCfgFile, err := filepath.Abs(cfgFile)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(absCfgFile, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	e := gob.NewEncoder(f)
	if err := e.Encode(i); err != nil {
		return err
	}
	return nil
}

func init() {
	appPath, _ := filepath.Abs(path.Dir(os.Args[0]))
	log.Printf("Find config in `%s`\n", appPath)
	if err := DefaultImages.ReadConfig(path.Join(appPath, "images.dat")); err != nil {
		log.Println(err)
	} else {
		log.Printf("Image data loaded.")
	}
}
