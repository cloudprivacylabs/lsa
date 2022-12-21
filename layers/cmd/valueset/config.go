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
	Databases []map[string]interface{} `json:"databases" yaml:"databases"`
}

func UnmarshalDatabasesConfig(configMap []map[string]interface{}, env map[string]string) (Config, error) {
	cfg := Config{
		ValuesetDBs: make([]ValuesetDB, 0),
		valueset:    make(map[string]ValuesetDB),
	}
	for _, rec := range configMap {
		for _, mp := range rec {
			for key := range mp.(map[string]interface{}) {
				if fn, ok := valuesetFactory[key]; ok {
					vsdb, err := fn(mp, env)
					if err != nil {
						return Config{}, err
					}
					cfg.valueset[key] = vsdb
					cfg.ValuesetDBs = append(cfg.ValuesetDBs, vsdb)
				}
			}
		}
	}
	return cfg, nil
}

func UnmarshalSingleDatabaseConfig(database map[string]interface{}, env map[string]string) (ValuesetDB, error) {
	for _, vals := range database {
		dbVals := cmdutil.YAMLToMap(vals)
		for key, value := range dbVals.(map[string]interface{}) {
			if fn, ok := valuesetFactory[key]; ok {
				vsdb, err := fn(value, env)
				if err != nil {
					return nil, err
				}
				return vsdb, nil
			}
		}
	}
	return nil, nil
}

// should get the root of the file
func LoadConfig(filename string, env map[string]string) (Config, error) {
	var uc unmarshalConfig
	if err := cmdutil.ReadJSONOrYAML(filename, &uc); err != nil {
		return Config{}, err
	}
	configMap := make([]map[string]interface{}, 0)
	for _, rec := range uc.Databases {
		for key, v := range rec {
			m := cmdutil.YAMLToMap(v)
			configMap = append(configMap, map[string]interface{}{key: m})
		}
	}
	return UnmarshalDatabasesConfig(configMap, env)
}
