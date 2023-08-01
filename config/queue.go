package config

type Queue struct {
	// do not process items means that we will just update their status in the DB
	// as if they were successfully processed (for testing purposes only!)
	SkipProcessing          bool `yaml:"skip_processing"`
	SkipBackroundProcessing bool `yaml:"skip_background_processing"`
}
