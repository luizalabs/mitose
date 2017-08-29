package config

type Config struct {
	Namespace   string `json:"namespace"`
	Deployment  string `json:"deployment"`
	Type        string `json:"type"`
	Max         int    `json:"max"`
	Min         int    `json:"min"`
	ScaleMethod string `json:"scale_method"`
}
