# prompt

## Overview

build a UI in the action/ directory. i need a layout with the same structure as workflow/main.go.

## Directives

1. lay out the model first and view then ask for review.
2. implement the data fetching in bubble tea messages.
3. refine the styles.

## Data

We'll need to fetch the install action workflow and the install action workflow runs. These should be modeled so that we
can display individual loading states for each.

the latest configured step needs to be easily accessible from the state.

## HEADER

the header should match the style of the header in `workflow/`. It should have a status indicator that is the status of
the latest run. the name of the action and action id should resemble the header in `workflow/`.

## FOOTER

and footer should be the same as the footer in workflow. a message component and the help section.

##

the middle section should have a list of runs on the left section and the Latest configured steps on the right. the
proportion should be switched. See the diagram below.

## Diagrams

```txt
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ action_workflow_name                                                                                                                    │
│ action_workflow_id                                                                                                                      │
│                                                                                                                                         │
├─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
├───────────────────────────────────────────────────────────────────────────────────────────────┐ ┌─────────────────────────────────────┐ │
│                                                                                               │ │Latest configured steps              │ │
│ Recent Executions                                                                             │ │┌───────────────────────────────────┐│ │
│                                                                                               │ ││              Step_1               ││ │
│ ┌───────────────────────────────────────────────────────────────────────────────────────────┐ │ │├───────────────────────────────────┤│ │
│ │action_name                                                                                │ │ ││Repository Details                 ││ │
│ │trigger                                                                 humanized time ago │ │ ││                                   ││ │
│ │run by: account@email.com                                                                  │ │ ││repo: ...                          ││ │
│ └───────────────────────────────────────────────────────────────────────────────────────────┘ │ ││branch: ...                        ││ │
│ ┌───────────────────────────────────────────────────────────────────────────────────────────┐ │ ││                                   ││ │
│ │action_name                                                                                │ │ │├───────────────────────────────────┤│ │
│ │trigger                                                                                    │ │ ││command                            ││ │
│ │run by: account@email.com                                                                  │ │ ││                                   ││ │
│ └───────────────────────────────────────────────────────────────────────────────────────────┘ │ ││codeblock                          ││ │
│                                                                                               │ │├───────────────────────────────────┤│ │
│                                                                                               │ ││variables                          ││ │
│                                                                                               │ ││                                   ││ │
│                                                                                               │ ││| foo | bar |                      ││ │
│                                                                                               │ ││| --- | --- |                      ││ │
│                                                                                               │ ││| foo | bar |                      ││ │
│                                                                                               │ ││                                   ││ │
│                                                                                               │ │└───────────────────────────────────┘│ │
│                                                                                               │ │┌───────────────────────────────────┐│ │
│                                                                                               │ ││              Step_2               ││ │
│                                                                                               │ │├───────────────────────────────────┤│ │
│                                                                                               │ ││Repository Details                 ││ │
│                                                                                               │ ││                                   ││ │
│                                                                                               │ ││repo: ...                          ││ │
│                                                                                               │ ││branch: ...                        ││ │
│                                                                                               │ ││                                   ││ │
│                                                                                               │ │├───────────────────────────────────┤│ │
│                                                                                               │ ││command                            ││ │
│                                                                                               │ ││                                   ││ │
│                                                                                               │ ││codeblock                          ││ │
│                                                                                               │ │├───────────────────────────────────┤│ │
│                                                                                               │ ││variables                          ││ │
│                                                                                               │ ││                                   ││ │
│                                                                                               │ ││| foo | bar |                      ││ │
│                                                                                               │ ││| --- | --- |                      ││ │
│                                                                                               │ ││| foo | bar |                      ││ │
│                                                                                               │ ││                                   ││ │
│                                                                                               │ │└───────────────────────────────────┘│ │
│                                                                                               │ │                                     │ │
│                                                                                               │ │                                     │ │
│                                                                                               │ │                                     │ │
│                                                                                               │ │                                     │ │
│                                                                                               │ │                                     │ │
│                                                                                               │ │                                     │ │
├───────────────────────────────────────────────────────────────────────────────────────────────┘ └─────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ >                                                                                                                                       │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```
