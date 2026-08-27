# MCP Tools

## App Branch Changes

I want to see an app branch, and what changed:

- "what changed in the most recent run of this app branch"
- "in the preview run for this PR, what changed" (looks up preview by latest pr)

## App Branch Preview

I want to run a preview of local code, against an app branch:

- "this should run something like nuon app branches preview" against my local dir
- "preview the changes in this pr against install foo"
- "follow the preview branch"

NOTE: this might need to use the local stdio to do things.

## App Branch Overview

Describe an app branch:

- "give me an overview of this app branch"
- "did this last run of this app branch succeed?"
- "describe what changed with this app branch and how far on the deployment it is"

## App Branch Run Progress

Work with an app branch run:

## App Overview

## MCP Context

This should use the local stdio, basically, you should be able to select a context. When creating a context, it should 
prompt you to create a token (read-only, app admin, etc). And then, you should be able to pin an org, app, install 
(optionally).

Any command should be able to pass the context in. The CLI should decorate all the calls  from the API tools with  this 
information, and pull this info out, so if you pass a context name "with this context" we pull it in and use it.

next, you should be able to pin a context in a session.
