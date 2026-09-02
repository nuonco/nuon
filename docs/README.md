# Nuon Docs

The public Nuon documentation, built with [Mintlify](https://mintlify.com). Navigation and
configuration live in `docs.json`.

### Development

Preview the docs locally with nuonctl:

```
nuonctl dev --dev=docs
```

Or run the dev server directly from this directory:

```
./dev.sh
```

Either way the preview is served at http://localhost:3333.

`dev.sh` pins the mint version to match `Dockerfile` and picks a node that mint
supports. mint refuses to run on node 25, which bun reports itself as, so
running `mint` under bun or an active node 25 fails. Overrides: `MINT_NODE`,
`MINT_VERSION`, `PORT`.

### Publishing Changes

Install our Github App to autopropagate changes from youre repo to your deployment. Changes will be deployed to production automatically after pushing to the default branch. Find the link to install on your dashboard.

#### Troubleshooting

- `mintlify is not supported on node 25` - `dev.sh` should avoid this; if it
  can't find a supported node, install one (`brew install node@24`) or point
  `MINT_NODE` at one.
- Page loads as a 404 - Make sure you are running in this directory, where
  `docs.json` lives.
