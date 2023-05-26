# assume role

This package allows you to quickly and easily assume an IAM role within code.

Throughout most of our services, we heavily lock down functionality and try to create dedicated roles for most things. Furthermore, all of our org infrastructure is isolated with IAM roles.


## Usage

To create a session and get credentials, do the following:

```go
v := validator.New()

assumer, err := assumerole.New(s.v,
  assumerole.WithRoleARN(s.AssumeRoleARN),
  assumerole.WithRoleSessionName(s.AssumeRoleSessionName))
if err != nil {
  return nil, fmt.Errorf("unable to create role assumer: %w", err)
}

cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
if err != nil {
  return nil, fmt.Errorf("unable to assume role: %w", err)
}

s3Client := s3.NewFromConfig(cfg)
```

You can also create an options function and pass it around, to make weaving parameters throughout code easier:

```go
opts := assumerole.Options{
  RoleARN: "role-arn",
  RoleSessionName: "session",
}

v := validator.New()

assumer, err := assumerole.New(s.v,
  assumerole.WithRoleARN(s.AssumeRoleARN),
  assumerole.WithRoleSessionName(s.AssumeRoleSessionName))
if err != nil {
  return nil, fmt.Errorf("unable to create role assumer: %w", err)
}
```
