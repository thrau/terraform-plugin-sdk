package retry

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

var StateChangeConfOverrides = map[string]StateChangeConf{}

var (
	cachedOverrides map[string]StateChangeConf
	cacheOnce       sync.Once
)

const overrideFileName = "terraform-retry-overrides.json"

func GetStateChangeConfOverrides() map[string]StateChangeConf {
	cacheOnce.Do(func() {
		loadOverrides()
	})

	// Return a copy to prevent modification of the cache
	overrides := make(map[string]StateChangeConf)
	for k, v := range cachedOverrides {
		overrides[k] = v
	}
	return overrides
}

func loadOverrides() {
	cachedOverrides = make(map[string]StateChangeConf)
	for k, v := range StateChangeConfOverrides {
		cachedOverrides[k] = v
	}

	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	path := cwd + "/" + overrideFileName
	println("Loading retry overrides from", path)
	if _, err := os.Stat(path); err == nil {
		overrides, err := LoadFromConfig(path)
		if err == nil {
			for k, v := range overrides {
				cachedOverrides[k] = v
			}
		}
	}
}

func LoadFromConfig(path string) (map[string]StateChangeConf, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var rawOverrides map[string]struct {
		Delay                     float64  `json:"Delay"`
		Pending                   []string `json:"Pending"`
		Target                    []string `json:"Target"`
		Timeout                   float64  `json:"Timeout"`
		MinTimeout                float64  `json:"MinTimeout"`
		PollInterval              float64  `json:"PollInterval"`
		NotFoundChecks            int      `json:"NotFoundChecks"`
		ContinuousTargetOccurence int      `json:"ContinuousTargetOccurence"`
	}
	if err := json.Unmarshal(data, &rawOverrides); err != nil {
		return nil, err
	}

	overrides := make(map[string]StateChangeConf)
	for k, v := range rawOverrides {
		overrides[k] = StateChangeConf{
			Delay:                     time.Duration(v.Delay * float64(time.Second)),
			Pending:                   v.Pending,
			Target:                    v.Target,
			Timeout:                   time.Duration(v.Timeout * float64(time.Second)),
			MinTimeout:                time.Duration(v.MinTimeout * float64(time.Second)),
			PollInterval:              time.Duration(v.PollInterval * float64(time.Second)),
			NotFoundChecks:            v.NotFoundChecks,
			ContinuousTargetOccurence: v.ContinuousTargetOccurence,
		}
	}

	return overrides, nil
}
