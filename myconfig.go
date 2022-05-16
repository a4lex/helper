package helper

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/a4lex/helper/myconfig"
)

const GENERAL_CONF_NAME = "general"

type myConfig struct {
	node       *myconfig.Node
	configName string
	inited     bool
}

var (
	CFG *myConfig = &myConfig{
		node:       nil,
		configName: "",
		inited:     false,
	}
	initConfigOnce sync.Once
)

// ConfigInit initializes the config
// @configFile: the config file path
// @configName: the name of the config in cconfig file
func ConfigInit(configFile, configName string) error {
	var err error = fmt.Errorf("config already inited")

	initConfigOnce.Do(func() {
		var f *os.File
		f, err = os.OpenFile(configFile, os.O_RDWR, 0660)
		if err == nil {
			var n *myconfig.Node
			parser := myconfig.NewParser(bufio.NewReader(f))
			if n, err = parser.Parse(); err == nil {
				CFG = &myConfig{n, configName, true}
			}
			f.Close()
		}
	})

	return err
}

// func (c *myConfig) StringVar(p *string, name string, value string) {

// }

// KeyExists returns true if the key exists in the config
// @key: the key of the node
func (c *myConfig) KeyExists(key string) bool {
	if _, err := c.value(key); err == nil {
		return true
	}

	return false
}

// String returns the string value of the node.
// If the node does not exist, it returns the default value.
// @key: the key of the node
// @defVal: the default value
func (c *myConfig) String(key, defVal string) string {

	if val, err := c.value(key); err == nil {
		return val
	}

	return defVal
}

// Int returns the integer value of the node.
// If the node does not exist, it returns the default value.
// @key: the key of the node
// @defVal: the default value
func (c *myConfig) Int(key string, defVal int) int {

	if val, err := c.value(key); err == nil {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}

	}

	return defVal
}

// Bool returns the boolean value of the node.
// If the node does not exist, it returns the default value.
// @key: the key of the node
// @defVal: the default value
func (c *myConfig) Bool(key string, defVal bool) bool {

	if val, err := c.value(key); err == nil {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}

	return defVal
}

// value returns the value of the node or error
// @key: the key of the node
func (c *myConfig) value(key string) (val string, err error) {

	if !c.inited {
		return "", fmt.Errorf("config not inited")
	}

	for _, fullKey := range []string{
		fmt.Sprintf("%s.%s", c.configName, key),
		fmt.Sprintf("%s.%s", GENERAL_CONF_NAME, key),
	} {

		val, err = c.node.GetNodeValue(fullKey)

		if err == nil || err != myconfig.ErrKeyNotFound {
			break
		}

	}

	return val, err
}

// func (c *myConfig) list(key string) (val []string, err error) {

// 	if !c.inited {
// 		return "", fmt.Errorf("config not inited")
// 	}

// 	for _, fullKey := range []string{
// 		fmt.Sprintf("%s.%s", c.configName, key),
// 		fmt.Sprintf("%s.%s", GENERAL_CONF_NAME, key),
// 	} {

// 		val, err = c.node.GetNodeListValues(fullKey)

// 		if err == nil || err != myconfig.ErrKeyNotFound {
// 			break
// 		}

// 	}

// 	return val, err
// }
