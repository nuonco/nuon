## Go Common

A shared library for go services to use


## Config subpackage example

 All the shown code would be in a service in this case example
 you can define a type and embed the base config struct with this annotation
 you can then set default values as well


 config.go:
```go
func init() {
    config.RegisterDefault("cache_redis_host", "localhost")
    config.RegisterDefault("cache_redis_port", uint16(6379))
    config.RegisterDefault("cache_redis_ttl", time.Minute*5)
    config.RegisterDefault("max_jobs_in_flight", 10000)
}

type Config struct {
    config.Base        `config:",squash"`
    GraphqlApiToken    string `config:"graphql_api_token"`
    ApiToken           string `config:"api_token"`
    RedisAddress       string `config:"redis_address"`
    MaxJobsInFlight    int    `config:"max_jobs_in_flight"`
    AwsSecretAccessKey string `config:"aws_secret_access_key"`
    AwsAccessKeyId     string `config:"aws_access_key_id"`
}
```

 If you don't want to set command line flags you do not have to but in this
 example we will

  main.go
```go
var rootCmd = &cobra.Command{
    Use: "example",
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println("we hit an error")
        os.Exit(2)
    }
}
```

 exec.go define a subcommand
```go
var execCmd = &cobra.Command{
      Use:   "exec",
      Short: "execs the program",
      Run:   exec,
 }
```

  exec.go: now we can use the init() function to initialize some defaults
```go
func init() {
    flags := execCmd.Flags()
    flags.String("service_name", "example", "the name of the service")
    flags.String("service_owner", "core", "the owner of the service")
    flag.String("service_part_of", "", "the name of a higher level application this service is a part of")
    rootCmd.AddCommand(execCmd)
}

```

 exec.go: and then the function the cmd calls
```go
func exec(cmd *cobra.Command, args []string) {
    var cfg Config
    if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
        log.Fatal("failed to load config", err.Error())
    }

    fmt.Printf("%#v", cfg)

}
```
 The library takes order of operations from Env -> Config file (json|yaml) -> cli flags
 you can use the default config.yml|json and set the values there or set environment varialbes
 that match the config: struct tags uppercased



