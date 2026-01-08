# App Branches

This is a feature which allows an app to track different versions of the same config, make changes and roll them out to 
different users.

The idea here is, like a branch, that an a user could open a PR modifying their nuon config, and then when they push to 
it, any install tied to that branch is updated. We will create a single workflow called an "app branch run" that will 
automatically fetch the config, build it, and then update each install for it.

## App Config Changes

The app config is currently built in the CLI, but will be moved to be built in the backend after this. We will retain 
the same sync package functionality, but use it from temporal.

## Queue Changes

We built queues for this use case to enable things like parallelization and what not. Please reference 
services/ctl-api/internal/pkg/queues for context here.

## Workflows for app

Currently we have install workflows. We will port that to use them in apps, where each update of an app branch will 
trigger the workflow for an install to be updated. Things like approvals and all that will be ported over.

Right now the workflows for stratus use a vertical style layout.

## App Branch Runs

An app branch run represents a single run of an app branch. This means fetching the config, and then building the 
components and then updating installs.

Most runs will come from a VCS connection commit. However, an app config can reference many different repos. We only let 
you define one per app branch, which is where the config is found. You can add regexes to ignore or allow files to 
trigger changes.

## CLI

We will build a TUI as part of this, and build all functionality we can into the CLI at the same time as regular.

## VCS Connection changes

We have some changes to the models that we will roll out to make vcs connections more a first class concept and have vcs 
commits and things like that. You can find some in the worktree commit here: /Users/jonmorehouse/nuon/worktrees/austin-app-branches

## Future Ideas

### Ability to dynamically create a branc

Create a branch per PR.
