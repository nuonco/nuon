# Flow: First-run onboarding

A brand-new account opens `/onboarding`, gets a trial org, and creates an app template plus a first test install.
Specs: `specs/onboarding/example-app.spec.ts`, `own-app.spec.ts`, `skip.spec.ts`, `resume.spec.ts`.

## Setup
- env: E2E_EMAIL (required)
- env: E2E_GITHUB_INSTALL_ID, E2E_ONBOARDING_REPO_AWS / _GCP / _AZURE (own-app specs; each skips when unset)
- note: each spec creates its own account (`e2e/onboarding.ts`) and ignores the global session
- note: ctl-api `force_sandbox_mode` must be on so the install stack reports back without a cloud account
- start: /onboarding

## Steps

### Intro, then Start
- action: goto | /onboarding
- expect: visible | heading "Your account is set up"
- expect: exactly one org exists for the account
- action: click | button "Create your first app template"
- expect: visible | heading "Create your first app template"

### Example app (example-app.spec.ts, AWS and GCP)
- expect: 4 stepper dots
- action: click | button "Deploy to AWS" (or "Deploy to GCP")
- expect: visible | heading "Your app is ready for BYOC"

### Own app (own-app.spec.ts, per cloud)
- setup: create the VCS connection with `POST /v1/vcs/connection-callback`
- action: goto | /onboarding?vcs-connected=<connection-id>
- action: click | button "Start with your app"
- expect: visible | text "Connected as"
- expect: 5 stepper dots
- action: fill | input "App template name" | My App
- expect: the name rule shows and Next is disabled; the value is unchanged
- action: fill | input "App template name" | <repo name>
- action: click | the cloud's test cloud tile
- action: click | button "Next"
- expect: visible | heading "Fill in your app template"
- expect: "See full prompt" links to https://nuon.co/loop.md
- action: click | button "Set up your first install"
- expect: visible | text "Nuon does not have your app config yet"
- action: click | button "Continue anyway"
- expect: visible | heading "Your app is ready for BYOC"

### Deploy to the workflow page (both paths)
- action: click | button "Create install"
- expect: visible | heading "Create the install stack" or "Your first BYOC install is deploying"
- action: if the stack is ready, launch it, then click | button "Continue"
- expect: visible | heading "Your first BYOC install is deploying"
- expect: visible | text "nuon installs workflows watch -i inl..."
- action: click | button "Go to deploy workflow"
- expect: url | /installs/<id>/history/<workflow-id>
- expect: every first_run step is complete

### Skip (skip.spec.ts)
- action: click | button "Skip"
- expect: url | /<orgId>
- expect: first_run complete, `skipped: "true"` on the start step
- action: goto | /
- expect: url | /<orgId>

### Resume (resume.spec.ts)
- action: click | button "Deploy to AWS"
- action: reload
- expect: visible | heading "Your app is ready for BYOC" with nuonco/kitchen-sink and nuonco/aws-eks-auto-sandbox
- action: click | button "Back"
- expect: visible | heading "Create your first app template"
- expect: start complete, deploy not complete
