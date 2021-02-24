package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
)

// Config the configuration for the app
type Config struct {
	// SNSTopicARN where
	SNSTopicARN string
	// WebServerAddr where the local webserver listens
	WebServerAddr string
	// Debug turn on debug logging
	Debug bool
}

// Lookup the env from |key| renamed to uppercase, hyphen is underscore, and return it or
// the |defaultVal| in the type of |defaultVal|
func lookupEnv(envNm string, defaultVal interface{}) (interface{}, error) {
	var res interface{}

	if val, ok := os.LookupEnv(envNm); ok {
		res = val
	} else {
		return defaultVal, nil
	}
	switch t := defaultVal.(type) {
	case int:
		v, err := strconv.Atoi(res.(string))
		if err != nil {
			return nil, fmt.Errorf("%w, lookupEnv[%s]: %v\n", err, envNm, res)
		}
		return v, nil
	case int64:
		v, err := strconv.ParseInt(res.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w, lookupEnv[%s]: %v\n", err, envNm, res)
		}
		return v, nil
	case float64:
		v, err := strconv.ParseFloat(res.(string), 64)
		if err != nil {
			return nil, fmt.Errorf("%w, lookupEnv[%s]: %v\n", err, envNm, res)
		}
		return v, nil
	case bool:
		bstr := strings.ToUpper(res.(string))
		if bstr == "TRUE" || bstr == "1" {
			return true, nil
		}
		return false, nil
	case string:
		return res, nil
	case time.Duration:
		v, err := time.ParseDuration(res.(string))
		if err != nil {
			return nil, fmt.Errorf("%w, lookupEnv[%s]: %v\n", err, envNm, res)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("lookupEnv[%s]: unsupported type %v", envNm, t)
	}
}

// ReadConfig loads config from command-line args (precedence) or environment
/* Example:
type MyConfig struct {
	FirstName string `tag_name:"tag 1"`
	LastName  string `tag_name:"tag 2"`
	Age       int    `tag_name:"tag 3"`
	Debug     bool   `tag_name:"debug"`		// specially-named field
}
*/
func ReadConfig(cfg interface{}) error {
	if err := readConfigWithFlagset(cfg, flag.CommandLine); err != nil {
		return err
	}
	flag.Parse()
	return nil
}

// a util to be able to use a different flagset
func readConfigWithFlagset(cfg interface{}, flagset *flag.FlagSet) error {
	if err := readConfig(cfg, flagset); err != nil {
		return err
	}
	return nil
}

func readConfig(cfg interface{}, flagset *flag.FlagSet) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("argument is not a struct pointer")
	}

	if err := reflectStruct(v, "", flagset); err != nil {
		return err
	}

	return nil
}

func reflectStruct(v reflect.Value, pfx string, flagset *flag.FlagSet) error {
	val := v.Elem()

	for i := 0; i < val.NumField(); i++ {
		fValue := val.Field(i)
		field := val.Type().Field(i)
		if !fValue.CanInterface() || !fValue.CanSet() {
			// Unexported struct field, can't reflect the interface.
			// Quietly ignore like json marshalling.
			continue
		}

		fTag := field.Tag

		// flag struct tag
		flagName := strcase.ToKebab(pfx) + strcase.ToKebab(field.Name)
		flagTag, flagTagOK := fTag.Lookup("flag")
		if flagTag != "" {
			if flagTag == "-" {
				// the ignore tag
				continue
			}
			flagName = strcase.ToKebab(pfx) + flagTag
		}

		// for a nested struct or struct pointer
		if fValue.Kind() == reflect.Ptr || fValue.Kind() == reflect.Struct {
			if fValue.Kind() == reflect.Ptr && fValue.IsNil() {
				continue
			}
			fpfx := flagName + "-"
			// an explicitly empty flagName on the nested structure means no prefix for its fields
			if flagTagOK && flagTag == "" {
				fpfx = ""
			}
			addr := fValue
			if fValue.Kind() != reflect.Ptr {
				addr = fValue.Addr()
			} else if addr.Elem().Kind() != reflect.Struct {
				continue
			}
			if err := reflectStruct(addr, fpfx, flagset); err != nil {
				return fmt.Errorf("%w; %s: field failure", err, field.Name)
			}
			continue
		}

		// env struct tag and default value
		defaultVal := fValue.Interface()
		envName := ""
		envTag, envTagOK := fTag.Lookup("env")
		if envTagOK {
			envName = envTag
		} else {
			envName = strcase.ToScreamingSnake(flagName)
		}
		// envTag of "-" means do not consider OS environment variable
		if envTag != "-" {
			d, err := lookupEnv(envName, defaultVal)
			if err != nil {
				return err
			}
			defaultVal = d
		}

		// usage struct tag
		flagUsage := fTag.Get("usage")

		if !fValue.CanAddr() {
			return fmt.Errorf("unable to address field %s", field.Name)
		}

		switch field.Type.String() {
		case "int":
			x := fValue.Addr().Interface().(*int)
			flagset.IntVar(x, flagName, defaultVal.(int), flagUsage)
		case "int64":
			x := fValue.Addr().Interface().(*int64)
			flagset.Int64Var(x, flagName, defaultVal.(int64), flagUsage)
		case "float64":
			x := fValue.Addr().Interface().(*float64)
			flagset.Float64Var(x, flagName, defaultVal.(float64), flagUsage)
		case "string":
			x := fValue.Addr().Interface().(*string)
			flagset.StringVar(x, flagName, defaultVal.(string), flagUsage)
		case "bool":
			x := fValue.Addr().Interface().(*bool)
			flagset.BoolVar(x, flagName, defaultVal.(bool), flagUsage)
		case "time.Duration":
			x := fValue.Addr().Interface().(*time.Duration)
			flagset.DurationVar(x, flagName, defaultVal.(time.Duration), flagUsage)
		default:
			return fmt.Errorf("unsuported struct type %s", field.Type.String())
		}
	}

	return nil
}
