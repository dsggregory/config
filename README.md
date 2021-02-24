# Command-line Config Loader

Easily manage program startup arguments with only the definition of a structure with special tags. Marshalling patterns fill in the rest. 

```shell
go get github.com/dsggregory/config
```

## Usage
Simply:
* define your `config` structure with optional struct tags
* create a variable of that structure type with defaults if desired
* call `ReadConfig()` with the address of that variable

The struct type defines the settings needed by your application. Only exported fields will be considered as command-line flag arguments.

### Take Care With Nested Structs
Only nested structs that are addressable (pointer) will be traversed. This means, unless you ignore the struct with the `flag:"-"` tag, then this will try to use command-line flags and environment variables to fill out your http.Client struct in your config (as an example).

### Supported Field Types
The common Golang flag types are supported:
* int
* int64
* float64
* string
* bool
* time.Duration

### Default Values & Precedence
Structure values at read-time are considered defaults, with corresponding but properly capitalized environment variable settings as a backup default.

Precedence of value choice follows:
* command-line flag
* environment variable
* struct value before call to ReadConfig()

### Struct Tags
Struct tags, quoted options following the field declaration, include the following along with their default capitalization style:

| Tag   | Description                                    | Style           |
|-------|------------------------------------------------|-----------------|
| flag  | command-line flag name. "-" means ignore. On nested structures, a value overrides the default prefix or an empty string prevents prefixing .    | field-name      |
| env   | environment variable name. "-" means ignore. Default env name is that of `flag`.  | FIELD_NAME      |
| usage | command-line flag usage                        |                 |

## Example

```go
type MyConfig struct {
	FirstName string `flag:"first_name" usage:"first name of user"`
	LastName  string `usage:"last name of user"`
	Age       int    `usage:"User's age in dog years"`
	Addr      Address `flag:""` // no prefix for flags of this struct
	Debug     bool   `env:"Debug"`
}

type Address struct {
	Street string `usage:"the street address"`
	Zip string `usage:"the postcode"`
}

cfg := MyConfig{
    FirstName: "DefaultFN",
    LastName:  "",
    Age:       0,
    Debug:     false,
}
err := ReadConfig(&cfg)
if err != nil {
	flag.PrintDefaults()
	os.exit(1)
}
```

Calling flag.PrintDefaults() from the above example will produce:
```text
  -age int
    	User's age in dog years
  -debug
    	
  -first_name string
    	first name of user (default "DefaultFN")
  -last-name string
    	last name of user
  -street
        the street address
  -zip
        the postcode
```
