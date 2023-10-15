package yamllib

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func DecodeByFile(yamlFile string, data any) error {
	fd, err := os.Open(yamlFile)
	if err != nil {
		return err
	}
	defer fd.Close()
	b, err := io.ReadAll(fd)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, data)
	if err != nil {
		return err
	}
	return nil
}
func DecodeByBytes(yamlContent []byte, data any) error {
	err := yaml.Unmarshal(yamlContent, data)
	if err != nil {
		return err
	}
	return nil
}
func Encode(data any) (string, error) {
	yaml, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(yaml), nil
}
