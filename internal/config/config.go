package config

import (
	"encoding/json"
	
	"path/filepath"

	"os"
)

const filename = ".gatorconfig.json"

type Config struct{
	DbUrl string `json:"db_url"`
	CurrentUser string `json:"current_user_name"`
}



func getConfigFilePath()(string,error){
	homePath, err := os.UserHomeDir()
	if err != nil{
		return "",err
	}
	fullPath := filepath.Join(homePath,filename)
	return fullPath, nil
}

func write(cfg Config)error{
	json, err := json.Marshal(cfg)
	if err != nil{
		return err
	}
	cfgPath, err := getConfigFilePath()
	if err != nil{
		return err
	}
	err = os.WriteFile(cfgPath,json,0633)
	if err != nil{
		return err
	}
	return nil
}





func Read()(Config,error){
	fullpath, err := getConfigFilePath()
	if err != nil{
		return Config{},err
	}

	content, err := os.ReadFile(fullpath)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(content,&config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}


func (c Config) SetUser(name string)error{
	c.CurrentUser = name
	err := write(c)
	if err != nil{
		return err
	}
	return nil
}	


