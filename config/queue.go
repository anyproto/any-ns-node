package config

type Queue struct {
	// if true - do not scan DB for items in "stuck" states
	// and do not try to process them
	SkipExistingItemsInDB bool `yaml:"skipExistingItemsInDB"`

	// do not process items means that we will just update their status in the DB
	// as if they were successfully processed (for testing purposes only!)
	SkipProcessing bool `yaml:"isSkipProcessing"`

	SkipBackroundProcessing bool `yaml:"isSkipProcessingBackground"`

	LowNonceRetryCount uint `yaml:"retryCountNonce"`

	HighNonceRetryCount uint `yaml:"retryCountHighNonce"`
}
