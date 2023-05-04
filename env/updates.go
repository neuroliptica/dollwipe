// updates.go: check for updates

package env

import (
	"dollwipe/network"
	"encoding/json"
	"fmt"
	"os"

	"github.com/neuroliptica/logger"
)

// Default logger for updates strout.
var UpdatesLogger = logger.MakeLogger("updates").BindToDefault()

// Links to remote repositories and remote manifest.
const (
	ManifestUrl = "https://raw.githubusercontent.com/neuroliptica/dollwipe/main/Manifest.json"
	Repository  = "https://github.com/neuroliptica/dollwipe"
)

// Manifest schema.
type Manifest struct {
	Version string `json:"version"`
	Updates string `json:"updates"`
}

// Unmarshal fetched bytes to manifest struct.
func ReadManifest(cont []byte) (*Manifest, error) {
	var manifest Manifest
	json.Unmarshal(cont, &manifest)
	if manifest.Version == "" {
		return nil, fmt.Errorf("invalid version, empty")
	}
	return &manifest, nil
}

// Get manifest from the remote origin.
func GetUpToDateManifest() (*Manifest, error) {
	cont, err := network.SendGet(ManifestUrl)
	if err != nil {
		return nil, err
	}
	return ReadManifest(cont)
}

// Get local manifest from manifest file.
func GetOldManifest(local string) (*Manifest, error) {
	cont, err := os.ReadFile(local)
	if err != nil {
		return nil, err
	}
	return ReadManifest(cont)
}

// Check for updates.
func FetchUpdates(local string) {
	failed := func(err error) {
		UpdatesLogger.Logf("ошибка получения: %v", err)
		UpdatesLogger.Log("не смогла проверить обновления, игнорирую.")
	}
	UpdatesLogger.Log("получаю свежий манифест...")
	fresh, err := GetUpToDateManifest()
	if err != nil {
		failed(err)
		return
	}
	UpdatesLogger.Logf("читаю локальный манифест... => %s", local)
	old, err := GetOldManifest(local)
	if err != nil {
		failed(err)
		return
	}
	if fresh.Version == old.Version {
		UpdatesLogger.Logf("версии совпадают [%s], обновлений нет.", fresh.Version)
		return
	}
	UpdatesLogger.Logf("нашла обновления: %s => %s", old.Version, fresh.Version)
	UpdatesLogger.Logf("изменения => %s", fresh.Updates)

	UpdatesLogger.Log("можешь обновить используя git: git pull && make build")
	UpdatesLogger.Logf("либо вручную отсюда: %s", Repository)
}
