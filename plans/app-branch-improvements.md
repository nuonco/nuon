## App Branch Improvements

### Diffs and Previews 

1. Stack needs to show diff when permissions change or when a customer input changes.

### PR Experience

1. Ignore drafts - make this an option, defaeult to true.
1. React when a preview run is triggered with a 👋 emoji.
1. Reuse the same comment, instead of making a new one.

### Sources

1. If a component is changed on the current preview, we need to overwrite it. For instance, if component tracks src/foo, 
   and that has changes, trigger a build with the current branch. If no changes, do not trigger this build.
1. Most of the time, the sandbox is not in the same repo. This workflow card can hang and not show what's actually going 
   on. We are not getting the source commit correctly configured, so it's not always getting the right commit or 
   caching.

### Many App Branches

1. If you have many app branches that are not completed, it's possible that you want to collapse the old ones. Let's 
   prototype out a UI to hanlde a situation where yotu have say 5 install groups, and 5-10 commits/runs and show how the 
   UI will present them. Say only the first group is auto-approved and the others are lingering back without being 
   modified.
1. There are some bugs where app branches just do not make progress when too many are outstanding.

### Install UI

1. Show the app branch run that the install is tracking + the commit message. make it easy to click into it. Think of it 
   as the "current app branch run"
1. Make the app branch update more useful - show if it was a preview, add a diff etc.
1. There should be a new overview page. This should show the app-branch run, whether its provisioned or not, and the 
   health status first.
1. The app branch run should be called "Updates". This should include. Each one should have a custom card.
- input changes
- stack changes
- install-config changes
- app branch run changes
1. History should be all workflows - should show manual deploys and scans there. - this should show workflows.

## Implementation

1. For all the UI changes, make new UI components and pages in ladle first, so I can see a real mockup.
1. For the app branches functionality, everything should work in the app-branch config file too.
