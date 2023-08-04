package config

type Queue struct {
	// if true - do not scan DB for items in "stuck" states
	// and do not try to process them
	SkipExistingItemsInDB bool `yaml:"skip_existing_items_in_db"`

	// do not process items means that we will just update their status in the DB
	// as if they were successfully processed (for testing purposes only!)
	SkipProcessing bool `yaml:"skip_processing"`

	SkipBackroundProcessing bool `yaml:"skip_background_processing"`
}
