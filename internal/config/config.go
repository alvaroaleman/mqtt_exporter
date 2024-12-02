package config

type Config struct {
	Configs map[string]SensorProperties `json:"configs"`
}

type SensorProperties struct {
	HumanReadableName string `json:"human_readable_name"`
}
