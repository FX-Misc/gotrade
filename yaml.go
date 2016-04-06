package gotrade

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func YamlFileDecode(path string, out interface{}) (err error) {
	// create file if not exists
	file, err := os.OpenFile(path, os.O_CREATE, 0644)
	if err != nil {
		return
	}
	file.Close()
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(content, out)
	if err != nil {
		return
	}
	return
}

func YamlFileEncode(path string, in interface{}) (err error) {
	out, err := yaml.Marshal(in)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path, out, 0644)
	if err != nil {
		return
	}
	return nil
}
