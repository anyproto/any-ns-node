package config

// Cache configures the Mongo name cache (see the "Name cache" section of the README)
type Cache struct {
	// the cache needs a replica set (or a sharded cluster): every write of a name is a
	// transaction that compares chain blocks. a standalone Mongo can not do that atomically,
	// two concurrent refreshes can then store an older chain state over a newer one.
	// the node refuses to start on a standalone Mongo unless this is set (local development only)
	AllowUnsafeStandalone bool `yaml:"allowUnsafeStandalone"`

	// how often the background repair scans the cache for the records that need a refresh
	// (seconds): incomplete ones and the re-reads after an operation. 0: the default (60), negative: off
	RepairIntervalSec int `yaml:"repairIntervalSec"`
}
