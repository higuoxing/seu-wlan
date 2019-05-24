package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
)

// Options ... Command line options
type Options struct {
	ConfigFile             string
	Username               string
	Password               string
	Interval               int
	Workers                int
	Timeout                int
	EnableMacAuth          bool
	DisableTLSVerification bool
}

// ReadFromConfigFile ... Read options from config file.
func ReadFromConfigFile(p string, o *Options) error {
	f, err := os.Open(p)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("Config File Parsing Error: %s", "cannot find your config file in this path")
	}

	content, err := ioutil.ReadAll(f)

	if err != nil {
		return fmt.Errorf("Config File Parsing Error: %s", "an error occurred when reading config file")
	}

	var configJSON map[string]interface{}
	err = json.Unmarshal(content, &configJSON)

	if err != nil {
		return fmt.Errorf("Config File Parsing Error: %s", "an error occurred when parsing config file")
	}

	if configJSON["username"] == nil || configJSON["password"] == nil {
		return fmt.Errorf("Config File Parsing Error: %s", "username and password are required")
	}

	switch ty := configJSON["username"].(type) {
	default:
		return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("username should be string format, not %T", ty))
	case string:
		o.Username = configJSON["username"].(string)
	}

	switch ty := configJSON["password"].(type) {
	default:
		return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("password should be string format, not %T", ty))
	case string:
		o.Password = configJSON["password"].(string)
	}

	// optional
	if configJSON["interval"] != nil {
		switch ty := configJSON["interval"].(type) {
		default:
			return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("interval should be integer, not %T", ty))
		case float64:
			o.Interval = int(configJSON["interval"].(float64))
		}
	}

	// optional
	if configJSON["enable-mac-auth"] != nil {
		switch ty := configJSON["enable-mac-auth"].(type) {
		default:
			return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("enable-mac-auth should be boolean, not %T", ty))
		case bool:
			o.EnableMacAuth = bool(configJSON["enable-mac-auth"].(bool))
		}
	}

	// optional
	if configJSON["disable-tls-verification"] != nil {
		switch ty := configJSON["disable-tls-verification"].(type) {
		default:
			return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("disable-tls-verification should be boolean, not %T", ty))
		case bool:
			o.DisableTLSVerification = bool(configJSON["disable-tls-verification"].(bool))
		}
	}

	// optional
	if configJSON["timeout"] != nil {
		switch ty := configJSON["timeout"].(type) {
		default:
			return fmt.Errorf("Config File Parsing Error: %s", fmt.Sprintf("disable-tls-verification should be boolean, not %T", ty))
		case bool:
			o.DisableTLSVerification = bool(configJSON["disable-tls-verification"].(bool))
		}
	}

	return nil
}

// VerifyOptions ... Verify config options.
func VerifyOptions(o *Options) error {
	if o.ConfigFile != "" {
		/* read from config file */
		err := ReadFromConfigFile(o.ConfigFile, o)
		if err != nil {
			return err
		}
	}

	if o.Username == "" || o.Password == "" {
		return fmt.Errorf("Command Parsing Error: %s", "username and password are required")
	} else if o.Interval < 0 {
		return fmt.Errorf("Command Parsing Error: %s", "-i option cannot be less than 0")
	} else if o.Workers <= 0 {
		return fmt.Errorf("Command Parsing Error: %s", "number of workers should be greater than 0")
	}

	return nil
}

// EncodePOSTForm ... Encode options into POST form.
func EncodePOSTForm(o *Options) url.Values {
	b64pass := base64.StdEncoding.EncodeToString([]byte(o.Password))

	var macAuth string

	if o.EnableMacAuth {
		macAuth = "1"
	} else {
		macAuth = "0"
	}

	return url.Values{
		"username":      {o.Username},
		"password":      {string(b64pass)},
		"enablemacauth": {macAuth},
	}
}
