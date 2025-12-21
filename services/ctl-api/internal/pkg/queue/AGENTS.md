# Queues

## Core Requirements

This package is the backbone of our distributed workflow tooling in the product. Different objects in our system are
backed by these, meaning that when an org, app, install or other is created we automatically spin up a queue to send
signals too.

Signals are workflows that do a user action, such as generating a plan, waiting for approvals, or running a deploy. It's
important that signals can be decoupled from each other and live in different directories. We have a registration system
where we use `init` functions to pick

Some requirements of the queue system:

1. we can send signals and see them persisted into the DB.
1. a signal can be retried, cancelled, etc and if temporal fails can be recovered.
1. we can run signals on a cron, so that a queue can have an "auto emitter" that sends a signal on a timeline.
1. we can use signals for control-flow. Since they are just workflows with queries, we can always ping a workflow and 
   the query will respond by restarting the loop and then returning the data from the db if it's not running.

The queue system has a nice client that we can use anywhere. It uses activities for things, so we want to be able to
embed it and expose as much functionality as we can in different places.

## Testing

We are building tests into the core system here, so that we can easily make sure the core queue tooling works properly
under different failure scenarios. I am trying to make sure that we also run the worker itself, vs just using the
temporal stubbing tooling.
