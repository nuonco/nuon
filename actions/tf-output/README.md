# action-tf-output

This actions takes a file containing the output of `terraform output -json` and
turns each tf output into an environment variable.

This is used to try to prevent hardcoding infrastructure knowledge in actions.
The infra is created / referenced by terraform so leverage its state to hold
the information. This action helps to parse the output into something more
usable in workflows.

```yml
    steps:
      - name: Get terraform outputs
        shell: bash
        working-directory: infra
        env:
          TF_WORKSPACE: stage
        run: |
          terraform init
          # The terraform setup action wraps the binary as `terraform`
          # So use the real binary for this output
          terraform-bin output -json > ${{ env.TF_OUT }}

      - name: Parse output
        id: tf_parse
        uses: powertoolsdev/action-tf-output@main
        with:
          file: ${{ env.TF_OUT }}

      # as an example, print all of the non-sensitive outputs
      - name: Echo output
        run: |
          printenv | grep TFO_
```

## How does it work

This is a typescript action that reads the input file and traverses the keys to
populate the environment.

Given an input like this:

```json
{
    "top_level":
    {
        "value": "something"
    },
    "also_top_level":
    {
        "value":
        {
            "nested": "somethingelse"
        }
    }
}
```

It will create 2 env vars - `TFO_TOP_LEVEL=something` and
`TFO_ALSO_TOP_LEVEL_NESTED=somethingelse`.

The terraform outputs are prefixed with `TFO_`, uppercased, and nested keys are
appended with `_`.


## Configuration

- `file`: A file containing the results of `terraform output -json`.
- `cleanup`: If true, removes the output json file. (default: `true`)
