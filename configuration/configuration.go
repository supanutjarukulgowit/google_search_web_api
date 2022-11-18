package configuration

import (
	"io/ioutil"

	"github.com/supanutjarukulgowit/google_search_web_api/util"
)

func LoadConfigFile(configPath string) (*Configuration, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	c := &Configuration{}

	err = util.ByteToStruct(data, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
