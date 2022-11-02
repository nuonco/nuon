# template-go-library

This is a template for creating a new go library.

## Usage

When you create a new repository in `infra-github` set the `template` field to `template-go-library`:

```terraform
module "my-go-library" {
  source = "./modules/repository"

  name          = "my-go-library"
  description   = "Go library"
  topics        = ["go-lib", ]
  from_template = "template-go-library"
}
```

Once you have created your repository, be sure to make an initial release and tag, named `v0.0.1`. This is required because each release of a library will look up the previous tag and use it to create the _next_ tag.

