# Pipeline

`pkg` Pipeline exposes a primitive for running pipelines of go functions, to compose steps of functions

## Background + why?

Often times, infrastructure code that is properly modularized requires building types that use dependency injection and then calling a set of functions on that type with side effects. If you look at a `terraform plan`, we need to initialize at least a handful of things:

* initialize backend
* initialize archive client
* unarchive
* run plan
* upload plan
* etc...

I originally wrote a package for doing this type of thing within the context of a cli, for `powertools`, and it worked out really well. By having a dedicated "pipeline" package, it allows you to keep types simple, exposing primitive functions and then composing them.

## Example usage

Without `pipeline`, you end up with lots of methods that simply compose smaller methods. For example, the following type of example demonstrates this:
```go

func (w  *Workspace) Load(ctx context.Context) error {
  if err := w.loadRoot(ctx); err != nil {
    return fmt.Errorf("unable to load root: %w", err)
  }

  if err := w.loadArchive(ctx); err != nil {
    return fmt.Errorf("unable to load root: %w", err)
  }

  if err := w.loadVariables(ctx); err != nil {
    return fmt.Errorf("unable to load variables: %w", err)
  }

  if err := w.loadBinary(ctx); err != nil {
    return fmt.Errorf("unable to load binary: %w", err)
  }
  if err := w.loadBackendc(ctx); err != nil {
    return fmt.Errorf("unable to load backend: %w", err)
  }
}

```

without pipeline, you have long methods that run many different things. Adding retry logic, printing out failures, adding dry-runs and more become extremely complicated. This also makes testing much, much harder and you end up with either a ton of additional interfaces, or tests that have many mocks.

With pipeline:

```go

func (w *Workspace) buildPipeline(ctx context.Context) (*pipeline.Pipeline, error) {
  pipe, err := pipeline.New(w.v)
  if err != nil {
    return nil, fmt.Errorf("unable to create pipeline: %w", err)
  }

  pipe.AddStep(&pipeline.Step{
    Name: "init root",
    Fn: w.initRoot,
  })

  pipe.AddStep(&pipeline.Step{
    Name: "init backend",
    Fn: w.initBackend,
  })

  pipe.AddStep(&pipeline.Step{
    Name: "init archive",
    Fn: w.initArchive,
  })

  ...
  return nil
}
```

This has a few benefits: it makes testing easier (you can test the pipeline without running it, to ensure ordering, steps), and it allows you to simplify your code.

## Pipelines vs temporal

This package is designed to be used outside of the context of `temporal`. While we use temporal to design workflows that call activities using the sdk, this package is designed for running pipelines of functions within a single

## Future use cases + roadmap

We plan on using this to implement different parts of our product, moving forward:

* executors - most executors run in a single activity
* helm plugin - our helm build / deploy plugin
* hooks plugin - helm build / deploy plugin

We plan on adding the following features:

* ability to share a log session between steps + upload output
* ability to retry steps
* custom mappers
* dry-run
