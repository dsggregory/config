package config

import (
	"flag"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/iancoleman/strcase"

	. "github.com/smartystreets/goconvey/convey"
)

func checkFlags(flags []*flag.Flag, names []string) bool {
	for _, flagName := range names {
		found := false
		for _, f := range flags {
			if f.Name == flagName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestConfig(t *testing.T) {
	// Your Config struct
	type MyConfig struct {
		FirstName string `flag:"first_name" usage:"first name of user"`
		LastName  string `usage:"last name of user"`
		Age       int    `usage:"User's age in dog years"`
		Swimmer   bool
		Debug     bool `env:"-"`
	}

	Convey("Spelling", t, func() {
		type casesT struct {
			val, expField, expEnv string
		}
		cases := []casesT{
			{"SnakeCase", "snake-case", "SNAKE_CASE"},
			{"Hyphen-Case", "hyphen-case", "HYPHEN_CASE"},
			{"Underscore_Case", "underscore-case", "UNDERSCORE_CASE"},
			{"FieldAPIKey", "field-api-key", "FIELD_API_KEY"},
			{"FieldURLAddr", "field-url-addr", "FIELD_URL_ADDR"},
			{"FieldHTTPServer", "field-http-server", "FIELD_HTTP_SERVER"},
		}
		for _, c := range cases {
			So(strcase.ToKebab(c.val), ShouldEqual, c.expField)
			So(strcase.ToScreamingSnake(c.val), ShouldEqual, c.expEnv)
		}
	})
	Convey("Basic", t, func() {
		firstName := "John"
		lastName := "Smith"
		agestr := "7"
		age := 7
		cfg := MyConfig{
			FirstName: firstName,
			LastName:  "",
			Age:       0,
		}

		os.Setenv("SWIMMER", "true") // default from env
		os.Setenv("DEBUG", "true")   // is ignored by env tag

		os.Args = []string{"cmd", "-last-name", lastName, "-age", agestr}
		err := ReadConfig(&cfg)
		//flag.PrintDefaults()
		So(err, ShouldBeNil)
		So(cfg.FirstName, ShouldEqual, firstName)
		So(cfg.LastName, ShouldEqual, lastName)
		So(cfg.Age, ShouldEqual, age)
		So(cfg.Debug, ShouldBeFalse)
		So(cfg.Swimmer, ShouldBeTrue)
	})

	Convey("Supported types", t, func() {
		type DurT struct {
			unexportedNotSet time.Duration
			Exported         time.Duration
			FloatT           float64
			IntT             int64
			Client           *http.Client `flag:"-"` // ignore
		}
		dur := DurT{Exported: time.Second, FloatT: 3.14159, IntT: 123456789, Client: &http.Client{}}
		os.Args = []string{"float-t", "3.14159", "int-t", "123456789", "timeout", "1h"}
		os.Args = []string{"cmd"}
		fs := flag.NewFlagSet("cmd", flag.ContinueOnError)
		err := readConfigWithFlagset(&dur, fs)
		So(err, ShouldBeNil)
		So(dur.Exported, ShouldEqual, time.Second)
		So(dur.FloatT, ShouldEqual, 3.14159)
		So(dur.IntT, ShouldEqual, 123456789)
		So(dur.Client.Timeout, ShouldNotEqual, time.Hour)
		flags := []*flag.Flag{}
		fs.VisitAll(func(f *flag.Flag) {
			flags = append(flags, f)
		})
		So(len(flags), ShouldEqual, 3) // 3 fields in total
		foundAll := checkFlags(flags, []string{"exported", "float-t", "int-t"})
		So(foundAll, ShouldBeTrue)
	})

	Convey("Env name based on flag", t, func() {
		type Ss1 struct {
			Zip string `flag:"postcode"`
		}
		ss := Ss1{}
		post := "w68rx"
		os.Setenv("POSTCODE", post)
		os.Args = []string{"cmd"}
		fs := flag.NewFlagSet("cmd", flag.ContinueOnError)
		err := readConfigWithFlagset(&ss, fs)
		So(err, ShouldBeNil)
		So(ss.Zip, ShouldEqual, post)
	})
	Convey("Nested structs", t, func() {
		type Ss2 struct {
			Street string
			Zip    string `flag:"postcode"`
		}
		type Ss1 struct {
			Name  string
			Addr  Ss2
			Addr2 *Ss2 `flag:""` // no nested prefix flag name
		}
		ss := Ss1{
			Addr2: &Ss2{}, // handles pointer to struct
		}
		street := "145 Hogarth Ln"
		post := "w68rx"
		os.Setenv("ADDR_STREET", street)
		os.Setenv("POSTCODE", post)
		os.Args = []string{"cmd"}
		fs := flag.NewFlagSet("cmd", flag.ContinueOnError)
		err := readConfigWithFlagset(&ss, fs)
		So(err, ShouldBeNil)
		So(ss.Addr.Street, ShouldEqual, street)
		So(ss.Addr2.Street, ShouldEqual, "")
		So(ss.Addr2.Zip, ShouldEqual, post)
		flags := []*flag.Flag{}
		fs.VisitAll(func(f *flag.Flag) {
			flags = append(flags, f)
		})
		So(len(flags), ShouldEqual, 5) // 5 fields in total
		foundAll := checkFlags(flags, []string{"name", "street", "postcode", "addr-street", "addr-postcode"})
		So(foundAll, ShouldBeTrue)
	})
}
