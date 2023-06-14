## confg

Nuon's opinionated approach to service configuration.

## Usage

By default, we have a `config.go` file in each service's `internal` directory. This config outlines the service's custom configuration and usually embeds either `workflows/worker/config.go` (worfklow worker) or `config.Base` from this package.


### Defaults

We consider defaults bad practice, and we should generally avoid them. Instead, set values for local in the services `service.yml` file.

When you need to define a default config, do so in the `init` function, as follows:

```go
func init() {
    config.RegisterDefault("default_key", "default_value")
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

To load the config in your service:

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
 you can use the default config.yml|json and set the values there or set environment variables
 that match the config: struct tags uppercased
