package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	//enc "github.com/hiromaily/golibs/cipher/encryption"
	u "github.com/hiromaily/golibs/utils"
)

var tomlFileName = os.Getenv("GOPATH") + "/src/github.com/hiromaily/go-job-search/data/settings.toml"

var conf *Config

// Config is of root
type Config struct {
	Matching *MatchingConfig `toml:"matching"`
	Keywords []KeywordConfig `toml:"keyword"`
	Page     *PagesConfig    `toml:"page"`
}

// MatchingConfig is for MySQL server
type MatchingConfig struct {
	Level uint8 `toml:"level"`
}

// KeywordConfig is for mail
type KeywordConfig struct {
	Search string `toml:"search"`
}

// PageConfig is for mail
type PagesConfig struct {
	Indeed        []PageConfig `toml:"indeed"`
	Stackoverflow []PageConfig `toml:"stackoverflow"`
	Linkedin      []PageConfig `toml:"linkedin"`
}

// PageIndeedConfig is for SMTP server of mail
type PageConfig struct {
	Url     string `toml:"url"`
	Param   string `toml:"param"`
	Country string `toml:"country"`
}

var checkTomlKeys = [][]string{
	{"matching", "level"},
}

//check validation of config
func validateConfig(conf *Config, md *toml.MetaData) error {
	//for protection when debugging on non production environment
	var errStrings []string

	//Check added new items on toml
	// environment
	//if !md.IsDefined("environment") {
	//	errStrings = append(errStrings, "environment")
	//}

	format := "[%s]"
	inValid := false
	for _, keys := range checkTomlKeys {
		if !md.IsDefined(keys...) {
			switch len(keys) {
			case 1:
				format = "[%s]"
			case 2:
				format = "[%s] %s"
			case 3:
				format = "[%s.%s] %s"
			default:
				//invalid check string
				inValid = true
				break
			}
			keysIfc := u.SliceStrToInterface(keys)
			errStrings = append(errStrings, fmt.Sprintf(format, keysIfc...))
		}
	}

	// Error
	if inValid {
		return errors.New("error: Check Text has wrong number of parameter")
	}
	if len(errStrings) != 0 {
		return fmt.Errorf("error: There are lacks of keys : %#v \n", errStrings)
	}

	return nil
}

// load configfile
func loadConfig(path string) (*Config, error) {
	if path != "" {
		tomlFileName = path
	}

	d, err := ioutil.ReadFile(tomlFileName)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading %s: %s", tomlFileName, err)
	}

	var config Config
	md, err := toml.Decode(string(d), &config)
	if err != nil {
		return nil, fmt.Errorf(
			"error parsing %s: %s(%v)", tomlFileName, err, md)
	}

	//check validation of config
	err = validateConfig(&config, &md)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// New is to create config instance
func New(file string, cipherFlg bool) *Config {
	var err error
	conf, err = loadConfig(file)
	if err != nil {
		panic(err)
	}

	//if cipherFlg {
	//	Cipher()
	//}

	return conf
}

// GetConf is to get config instance. singleton architecture
func GetConf() *Config {
	var err error
	if conf == nil {
		conf, err = loadConfig("")
	}
	if err != nil {
		panic(err)
	}

	return conf
}

// SetTOMLPath is to set TOML file path
func SetTOMLPath(path string) {
	tomlFileName = path
}

// ResetConf is to clear config instance
func ResetConf() {
	conf = nil
}

// Cipher is to decrypt crypted string on config
//func Cipher() {
//	crypt := enc.GetCrypt()
//}
