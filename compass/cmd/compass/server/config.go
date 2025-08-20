package server

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/ossf/gemara/layer2"

	compass "github.com/complytime/complybeacon/compass/service"
	"github.com/complytime/complybeacon/compass/transformer"
	"github.com/complytime/complybeacon/compass/transformer/factory"
)

func NewScopeFromCatalogPath(catalogPath string) (compass.Scope, error) {
	cleanedPath := filepath.Clean(catalogPath)
	catalogData, err := os.ReadFile(cleanedPath)
	if err != nil {
		return nil, err
	}

	var layer2Catalog layer2.Catalog
	err = yaml.Unmarshal(catalogData, &layer2Catalog)
	if err != nil {
		return nil, err
	}

	return compass.Scope{
		layer2Catalog.Metadata.Id: layer2Catalog,
	}, nil
}

type Config struct {
	Plugins     []PluginConfig `json:"plugins"`
	Certificate CertConfig     `json:"certConfig"`
}

type CertConfig struct {
	PublicKey  string `json:"cert"`
	PrivateKey string `json:"key"`
}

type PluginConfig struct {
	Id             string `json:"id"`
	EvaluationsDir string `json:"evaluations-dir"`
}

// TODO: This need to be easier to fallback to more generic processing

func NewTransformerSet(config *Config) (transformer.Set, error) {
	pluginSet := make(transformer.Set)
	for _, pluginConf := range config.Plugins {
		transformerId := transformer.ID(pluginConf.Id)
		if pluginConf.EvaluationsDir == "" {
			log.Printf("Plugin %s has no evaluations, skipping...", transformerId)
			continue
		}

		info, err := os.Stat(pluginConf.EvaluationsDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return pluginSet, fmt.Errorf("evaluations directory %s for plugin %s: %w", pluginConf.EvaluationsDir, pluginConf.Id, err)
			}
			return pluginSet, err
		}

		if !info.IsDir() {
			return pluginSet, fmt.Errorf("evaluations directory %s for plugin %s is not a directory", pluginConf.EvaluationsDir, pluginConf.Id)
		}

		tfmr, err := NewTransformerFromDir(transformerId, pluginConf.EvaluationsDir)
		if err != nil {
			return pluginSet, fmt.Errorf("unable to load configuration for %s: %w", pluginConf.Id, err)
		}
		pluginSet[transformerId] = tfmr
	}
	return pluginSet, nil
}

func NewTransformerFromDir(pluginID transformer.ID, evaluationsPath string) (transformer.Transformer, error) {
	tfmr := factory.TransformerByID(pluginID)
	err := filepath.Walk(evaluationsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var evaluation transformer.EvaluationPlan
		err = yaml.Unmarshal(content, &evaluation)
		if err != nil {
			return err
		}

		tfmr.AddEvaluationPlan(evaluation)
		return nil
	})
	return tfmr, err
}
