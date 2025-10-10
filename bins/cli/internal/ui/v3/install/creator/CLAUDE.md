prompt

this directory is for an install creation TUI. this is a run-of-the-mill contextual TUI using bubbletea.

## styles

- Use the styles from the `pkg/cli/styles` in a manner consistent with the way styles are used in the `workflow/` and
  `action/` directories.
- Use the bubble tea forms package.

## Challenge

The big challenge here is going to be dynamically composing a form. The form should be dynamically generated.

## Work plan

1. fetch the app config and read the inputs.
2. use the inuts to dynamically generate a form.

in addition to the dynamically generated fields we need to standard fields.

1. name
2. region (us aws regions)

Form submission shoud use the CreateInstall method from nuon-go.
