package config

import _ "embed"

//go:embed data/certifications.yaml
var certificationsYAML []byte

type CertificationType struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
}

type CertificationsConfig struct {
	Categories     []Category          `yaml:"categories"`
	Certifications []CertificationType `yaml:"certifications"`
}

var AllCertificationTypes []CertificationType
var AllCertificationCategories []Category
