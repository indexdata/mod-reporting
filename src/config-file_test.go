package main

import "testing"
import "gotest.tools/assert"


func Test_readConfig(t *testing.T) {
	t.Run("non-existent config file", func(t *testing.T) {
		cfg, err := readConfig("/no/such/file.json")
		assert.Error(t, err, "open /no/such/file.json: no such file or directory")
		var nilConfig *config
		assert.Equal(t, cfg, nilConfig)
	})

	t.Run("well-known config file", func(t *testing.T) {
		cfg, err := readConfig("../etc/config.json")
		assert.NilError(t, err)
		assert.DeepEqual(t, cfg, &config{
			Logging: loggingConfig{
				Categories: "config,listen,path,error",
			},
			Listen: listenConfig{
				Host: "0.0.0.0",
				Port: 12369,
			},
		})
	})
}
