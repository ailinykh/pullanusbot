package main

import (
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_ServerConfig_ReadsConfigFromBase64String(t *testing.T) {
	data := []byte(`
	[
		{
			"key_id": "some_key_id",
			"secret_key": "secret",
			"chat_ids": [1020304050607080, -1122334455667788],
			"command": "/reboot"
		}
	]
	`)

	filePath := path.Join(os.TempDir(), uuid.NewString()+".json")
	err := os.WriteFile(filePath, data, 0644)
	assert.Nil(t, err)

	configs, err := NewServerConfigList(filePath)
	assert.Len(t, configs, 1)

	for _, config := range configs {
		assert.Nil(t, err)
		assert.Equal(t, config.GetKeyID(), "some_key_id")
		assert.Equal(t, config.GetSecretKey(), "secret")
		assert.Equal(t, config.GetChatIds(), []int64{1020304050607080, -1122334455667788})
		assert.Equal(t, config.GetCommand(), "/reboot")
	}
}
