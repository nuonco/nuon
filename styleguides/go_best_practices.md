# Go best practices

This document is a WIP outlining best practices that we use throughout our codebase.

## wrap all errors

## validate structs wherever possible

## functional options + validation

We prefer functional options, and validate parameters into structs. This is designed to "shift left" configuration errors so that we can identify any issues with a piece of code during initialization.

Most non-initialization code has unit tests, so between this pattern and our unit testing we have a pretty reliable way of making changes.

## pkg first driven design

## prefer small, composable packages

Prefer smaller packages that compose each other. This makes it easier to test things in isolation, and move towards a place where something is reusable.

## packages define interfaces, return concrete types

* return solid types for LSP
* interfaces defined so we don't have to "wrap" things
