# App Branch Previews

## Goal

Enable a developer to push a branch, and see a quick preview of the changes they are making on their current install.

## Approach

Previously, we tried to overrun the app-branch-run workflow. The idea was that you'd mark an install group as a 
"preview" install group. However, this has a few issues:

1. it makes it hard to handle an install belonging to different branches. For instance, if you have an install that is 
   on a main branch and deploy to it via preview on that branch, you have to run for the whole branch and have a lot of 
   leftover installs.
2. it is hard to create a "one off" branch run. For instance, i want to be able to click "preview" against a branch, and 
   see hte changes + pick an install. The api I'm imaging let's you pass in a git-branch (to fetch head), an app-branch 
   (to base on), and one or more installs to preview against. If you don't pass any installs in to preview, then we just 
   do a build and throw the app-config away.

There's a few things I'll want to solve in here as well:
1. I want to make sure app-configs get labels (look at hte labels gorm type we have.) we should classify the app-configs 
   as manual, branch-run or preview.
2. We should make sure that the app-branch-runs that are a preview, show up as previews under that branch. We should 
   have some collapsing in there. For instance, if I pass in and say "preview against branch main", the branch-run that 
   we create should show it's a clear preview, it should be filterable out in the UI, and we should be able view them 
   within the main app branch.

## Implementation

1. let's build the app branch run workflow and make it so it can take a preview run in. The app branch preview should be 
   a new object that links to the parent. Like an app-branch-run-preview-config -- has a "build-only flag", a "pr-flag" 
   (to control what PR to post back too), and an "install" flag with "install plan-only" or "install-deploy" option.
2. the app branch run, if it has a preview, the steps will generate differently. Maybe we have a different workflow, but 
   that might be overkill since it's a similar object.
3. We should remove the preview option from teh app branch config - the preview option should not be on the 
   install0groups at all, so can be removed.
4. Let's add a new command called "nuon apps branches preview"
    - let's you pick a remote branch, a pr, or your local code
    - let's you pick a target app branch
    - let's you pick an install from that app branch
    - let's you pick whether to build only, deploy, or plan-only-deploy
    - opens the workflow for you locally that runs the app-branch
5. next, if a github event is a branch push, but to a PR one - it should do a preview run, but _only_ build-only.
6. we should have the run view in the list view collapse previews, and have an option to filter out previews.
7. mcp tool - you should be able to tell your agent to preview this with an install, and then listen for changes on the 
   workflow by polling it.
