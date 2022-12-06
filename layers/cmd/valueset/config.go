package valueset

import (
	"context"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
)

type Config struct {
	ValuesetDBs []ValuesetDB
	valueset    map[string]ValuesetDB
}

func (cfg Config) ValueSetLookup(ctx context.Context, tableId string, env map[string]string) (map[string]string, error) {
	for _, vsb := range cfg.ValuesetDBs {
		tIds := vsb.GetTableIds()
		if tIds != nil {
			if _, ok := tIds[tableId]; ok {
				return vsb.ValueSetLookup(ctx, tableId, env)
			}
		}
	}
	return nil, nil
}

type ValuesetDB interface {
	ValueSetLookup(context.Context, string, map[string]string) (map[string]string, error)
	GetTableIds() map[string]struct{}
	Close() error
}

var valuesetFactory = make(map[string]func(interface{}, map[string]string) (ValuesetDB, error))

func RegisterDB(name string, fn func(interface{}, map[string]string) (ValuesetDB, error)) {
	valuesetFactory[name] = fn
}

type unmarshalConfig struct {
	Databases map[string]interface{} `json:"databases" yaml:"databases"`
}

func UnmarshalConfig(configMap map[string][]interface{}, env map[string]string) (Config, error) {
	cfg := Config{
		ValuesetDBs: make([]ValuesetDB, 0),
		valueset:    make(map[string]ValuesetDB),
	}
	for _, rec := range configMap {
		for _, mp := range rec {
			for _, vals := range mp.(map[string]interface{}) {
				for key := range vals.(map[string]interface{}) {
					if fn, ok := valuesetFactory[key]; ok {
						vsdb, err := fn(vals, env)
						if err != nil {
							return Config{}, err
						}
						cfg.valueset[key] = vsdb
						cfg.ValuesetDBs = append(cfg.ValuesetDBs, vsdb)
					}
				}
			}

		}
	}
	return cfg, nil
}

func LoadConfig(filename string, env map[string]string) (Config, error) {
	var uc unmarshalConfig
	if err := cmdutil.ReadJSONOrYAML(filename, &uc.Databases); err != nil {
		return Config{}, err
	}
	ymlMap := make(map[string][]interface{}, 0)
	for key, rec := range uc.Databases {
		m := cmdutil.YAMLToMap(rec)
		ymlMap[key] = m.([]interface{})
	}
	return UnmarshalConfig(ymlMap, env)
}
