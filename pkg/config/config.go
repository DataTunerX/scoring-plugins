package config

import "github.com/spf13/viper"

var config *viper.Viper

func init() {
	config = viper.New()
	config.AutomaticEnv()
	config.BindEnv("level", "LOG_LEVEL")
	config.SetDefault("level", "debug")
	// bind COMPLETE_NOTIFY_URL env var
	config.BindEnv("complete_notify_url", "COMPLETE_NOTIFY_URL")
	// bind DATATUNERX_SYSTEM_NAMESPACE env var
	config.BindEnv("datatunerx_system_namespace", "DATUNERX_SYSTEM_NAMESPACE")

}

func GetLevel() string {
	return config.GetString("level")
}

// GetCompleteNotifyURL fetch COMPLETE_NOTIFY_URL env var
func GetCompleteNotifyURL() string {
	return config.GetString("complete_notify_url")
}

// GetDatatunerxSystemNamespace fetch DATUNERX_SYSTEM_NAMESPACE env var
func GetDatatunerxSystemNamespace() string {
	return config.GetString("datatunerx_system_namespace")
}
