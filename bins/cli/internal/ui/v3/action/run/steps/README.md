# Action Workflow Step List Component

## Attributes

- width: passed down from calling context
- height: passed down from calling context
- ActionWorkflowRun: passed down from calling context
- Steps: managed internally, munged from ActionWorkflowRun as necessary. Stored in a
  `map[string]AppInstallActionWorkflowRunStep{}` where the string is the step's name.
- LogStream: Found in ActionWorkflowRun.RunnerJob.LogStreamID
- Logs: fetched via logstream in pages. stored in a `map[string]AppOtelLogRecord` where the string is the value from
  each `AppOtelLogRecord`'s `LogAttributes`'s `workflow_step_name`. This way we can filter the logs by
  workflow_step_name when we want to display a single step's logs. Logs and the LogStream are fetched periodically,
  every five seconds. If the logstream is closed, we fetch the next page and then stop fetching.

## View

The component's size is determined externally and it receives its sizes instead of determining itself.

## Data

- Top level object is an ActionWorkflowRun
- The Steps are children of the ActionWorkflowRun.
- The logs are fetched for each step when the step is opened.
- Logs are stored in an array for each step so they do not need to be refetched.
- If the logstream is open, we fetch the next page every 5s.
- If the logstream is closed, we fetch all of the pages and idle.

## Diagram

```
┌────────────────┐
│   Steps App    │
└────────────────┘
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                              │
│steps                                                                                         │
│                                                                                              │
│┌───────────────────────────────────────────────────────────────────────────────────────────┐ │
││                                                                                           │ │
││[s] step.name                                                                              │ │
││                                                                                           │ │
│└───────────────────────────────────────────────────────────────────────────────────────────┘ │
│┌───────────────────────────────────────────────────────────────────────────────────────────┐ │
││                                                                                           │ │
││[s] step.name                                                                              │ │
││                                                                                           │ │
│└───────────────────────────────────────────────────────────────────────────────────────────┘ │
│┌───────────────────────────────────────────────────────────────────────────────────────────┐ │
││                                                                                           │ │
││[s] step.name                                                                              │ │
││                                                                                           │ │
│└───────────────────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                              │
│                                                                                              │
│                                                                                              │
│                                                                                              │
│                                                                                              │
│                                                                                              │
│                                                                                              │
│                                                                                              │
│                                                                                              │
│                                                                                              │
│                                                                                              │
│                                                                                              │
└──────────────────────────────────────────────────────────────────────────────────────────────┘

┌───────────────────────┐
│  Steps App: Expanded  │
└───────────────────────┘
┌───────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                               │
│[s] step.name                                                                                  │
│                                                                                               │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│LEVEL MESSAGE TIMESTAMP                                                                        │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│LEVEL MESSAGE TIMESTAMP                                                                        │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│LEVEL MESSAGE TIMESTAMP                                                                        │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│LEVEL MESSAGE TIMESTAMP                                                                        │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│LEVEL MESSAGE TIMESTAMP                                                                        │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│LEVEL MESSAGE TIMESTAMP                                                                        │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│LEVEL MESSAGE TIMESTAMP                                                                        │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│LEVEL MESSAGE TIMESTAMP                                                                        │
├───────────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                               │
│                                                                                               │
│                                                                                               │
│                                                                                               │
│                                                                                               │
│                                                                                               │
│                                                                                               │
│                                                                                               │
│                                                                                               │
│                                                                                               │
└───────────────────────────────────────────────────────────────────────────────────────────────┘
```
