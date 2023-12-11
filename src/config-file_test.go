package main

import "testing"
import "gotest.tools/assert"


func Test_readConfig(t *testing.T) {
	t.Run("non-existent config file", func(t *testing.T) {
		cfg, err := readConfig("/no/such/file.json")
		var nilConfig *config
		assert.Equal(t, cfg, nilConfig)
		assert.Error(t, err, "open /no/such/file.json: no such file or directory")
	})

	t.Run("malformed config file", func(t *testing.T) {
		cfg, err := readConfig("../etc/not-json.txt")
		var nilConfig *config
		assert.Equal(t, cfg, nilConfig)
		assert.Error(t, err, "invalid character 'T' looking for beginning of value")
	})

	t.Run("well-known config file", func(t *testing.T) {
		cfg, err := readConfig("../etc/silent.json")
		assert.NilError(t, err)
		assert.DeepEqual(t, cfg, &config{
			Logging: loggingConfig{},
			Listen: listenConfig{
				Host: "0.0.0.0",
				Port: 12369,
			},
		})
	})
}
