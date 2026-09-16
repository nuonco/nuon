## App Branch Version Management

How to manage installs, when different customers have different versions, release cadences and update roll outs.

> We propose a model where an app branch is created per end-customer (tracking main), and has install-groups to control 
> updates between dev and prod installs. Tags can be used to control which app branches are run, allowing you to have a 
> single trunk, but also support customers that have their own release cadences. Branch overrides allow you to easily 
> set overrides on a branch.

### Version Management

For BYOC Nuon, app branches do not work natively. One of the challenges that is core is that not _every_ single change 
should be deployed to all customers. In a git-centric world, what we really care about is having a view between the 
customer's desired vs actual state, but in BYOC some customers might be on release windows.

For instance, let's say you have 10 customers:

- Group 1 - 5 customers are "GA" or standard tier, they simply care that their install is updated when cloud is updated.
- Group 2 - 5 customers have different release cadences. They want to be on specific app config versions, at different 
  points in time, and only update when you have worked with them.

There's two approaches here, that make sense:

1. You could make the application version an input, and every change to the app is deployed to _all customers_. This 
   does not allow the customers to remove permissions, and means that all application changes must be decoupled from 
   infra changes.
2. You allow a customer install to track a branch, but can remove it from the branch between updates. This allows you to 
   stay at a single point in time of the app-config. If you need to apply a hot-fix, you can create a branch for that 
   customer from the commit they track and then push changes.

Let's also complicate things by this flow - let's consider that only some install groups are sequential. For instance, 
let's say group 1 has the first five "standard" installs, but hten group 2 has the five installs that are all completely 
separately managed. You might have a situation where we update the first 5, then want to update one of the second group. 
HOwever, the plans could go stale, they could be changed at different times etc.

We need tooling for re-running a plan, etc.

### Subsequent Run Issue

Now, let's exacerbate the problem further. Let's say you push 3 commits to an app branch tracking main. Let's walk 
through each one:

Push 1 - you make an app change. It's auto applied to the first group. No customer has given approval so the second 
group is ignored.
Push 2 - you make an infra change. It's auto applied to first group (again). Only one customer has given approval, so 
now you want to approve only a single install in the second group.
Push 3 - you make both an app + infra change. Auto applied again. 

Now, when you go to the app branch overview. You want to see a few things:

1. you want to see that the first group is up to date with the app branch's latest run.
2. you want to see the one install in the second group is up to date with commit 2.
3. you should see the other installs are behind.

What is the correct mechanism? Should the old installs just remain lingering? Should we have rules for allowing you to 
only selectively opt into new changes?

### Tags

Not all commits are the same. For instance, some groups should track all commits on a git branch. Other groups, maybe 
they should only track commits that have a tag.

I am thinking about the use case of managing customer's. It seems to me like we could do something where:

1. Group 1 - gets updated on all commits.
2. Group 2 - represents only a single of the five customers. You declare a tag ref that is used to update it. When a tag 
   is pushed, that is pushed to the install-group. For instance, you could have tag `customer-foo-v0.0.1` where foo is 
   there name. Then, each customer gets it's own group.

NOTE: you could also have these on the branch level. this would mean that you could have a branch for each customer, and 
declare the tag format that is used for each push to that branch. That means you could create a branch per customer, and 
connect installs together.

Let's say that customer "foo", has a dev, stage and prod instance. you could define a group for each, and then the 
branch only is run when a tag is imported.

> NOTE: if we had tags defined at the install-group level, the challenge is that a single app branch for everyone would 
> have a lot of "noise". you'd have many (at scale) tag pushes for specific customers that still result in a build/plan, 
> vs just being out-right skipped.

### Run Configuration

Run cadence is configured on the app branch:

```toml
[run]
mode = "on_tag_prefix"
tag_prefix = "foobar/"
```

Supported modes:

- `all` (default): run tracked-branch pushes and PR previews.
- `on_tag_prefix`: run a matching tag only when its commit is on the tracked branch's history.
- `on_github_label`: after a PR lands, run when its GitHub labels contain the configured exact label.
- `manual_only`: never run automatically.

Every run stores a JSONB provenance snapshot on `AppBranchRun.metadata`. `run_type` remains the execution shape
(`git-run`, `git-preview-run`, or `manual-run`), while metadata records why it ran (push, PR, tag, GitHub label, or
manual), the SHA/ref, PR number, and the branch run mode that matched. Status metadata is not the source of truth for
run identity.

### Overrides

We currently support install overrides, which allow you to declare overrides of a component, the stack, or sandbox at 
the install level.

Since many groups of installs will be connected together, and deployed in their own rules, it follows that we should 
allow overrides. For instance, if you have a branch per customer thta follows it's own tag, then you probably want to be 
able to set some overrides for all installs in that branch.
