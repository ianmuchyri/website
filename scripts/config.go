package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func GetDefaultConfig() *Config {
	return &Config{
		Site: SiteConfig{
			Name: "Abstract Machines",
			Url:  "https://absmach.eu",
		},
		Blog: BlogConfig{
			DateFormat:   "January 02, 2006",
			ReadingSpeed: 200,
			CategoryColors: map[string]string{
				"blog": "primary",
			},
		},
		Theme: ThemeConfig{
			PrimaryColor: "#073763",
			FontFamily:   "Rubik, sans-serif, fallback for Rubik",
			CodeTheme:    "github-dark",
		},
	}
}
