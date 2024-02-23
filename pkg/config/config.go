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
	config.BindEnv("datatunerx_system_namespace", "DATATUNERX_SYSTEM_NAMESPACE")
	// bind IN_TREE_SCORING_IMAGE env var
	config.BindEnv("in_tree_scoring_image", "IN_TREE_SCORING_IMAGE")
	config.BindEnv("datatunerx_server_name", "DATATUNERX_SERVER_NAME")
	config.SetDefault("datatunerx_server_name", "datatunerx-server")
	config.BindEnv("rouge1_weight", "ROUGE1_WEIGHT")
	config.SetDefault("rouge1_weight", "0.35")
	config.BindEnv("rouge2_weight", "ROUGE2_WEIGHT")
	config.SetDefault("rouge2_weight", "0.4")
	config.BindEnv("rougeL_weight", "ROUGEL_WEIGHT")
	config.SetDefault("rougeL_weight", "0.15")
	config.BindEnv("rougeLsum_weight", "ROUGELSUM_WEIGHT")
	config.SetDefault("rougeLsum_weight", "0.1")
	config.BindEnv("rouge_weight", "ROUGE_WEIGE")
	config.SetDefault("rouge_weight", "0.75")
	config.BindEnv("bleu_weight", "BLEU_WEIGHT")
	config.SetDefault("bleu_weight", "0.25")
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

// GetInTreeScoringImage fetch IN_TREE_SCORING_IMAGE env var
func GetInTreeScoringImage() string {
	return config.GetString("in_tree_scoring_image")
}

func GetDatatunerxServerName() string {
	return config.GetString("datatunerx_server_name")
}

func GetRouge1Weight() string {
	return config.GetString("rouge1_weight")
}

func GetRouge2Weight() string {
	return config.GetString("rouge2_weight")
}

func GetRougeLWeight() string {
	return config.GetString("rougeL_weight")
}

func GetRougeLsumWeight() string {
	return config.GetString("rougeLsum_weight")
}

func GetRougeWeight() string {
	return config.GetString("rouge_weight")
}

func GetBleuWeight() string {
	return config.GetString("bleu_weight")
}
