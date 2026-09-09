import { useState } from 'react'
import { DateTime } from 'luxon'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { Badge } from './Badge'
import { Banner, type TBannerTheme } from './Banner'
import { Button } from './Button'
import { Text } from './Text'

export default {
  title: 'lite/atoms/Banner',
}

const THEMES: Array<{
  theme: TBannerTheme
  heading: string
  body: string
}> = [
  {
    theme: 'default',
    heading: 'Config sync is enabled',
    body: 'Settings are pulled from the install config file on every deploy.',
  },
  {
    theme: 'success',
    heading: 'Configuration updated',
    body: 'The new configuration applies to the next deploy.',
  },
  {
    theme: 'error',
    heading: 'Deploy failed',
    body: 'Unable to deploy payments to production. This is usually temporary.',
  },
  {
    theme: 'warn',
    heading: 'Runner version is out of date',
    body: 'Upgrade the runner to keep deploys running.',
  },
  {
    theme: 'info',
    heading: 'Deploy in progress',
    body: 'Deploying payments to production. This may take a few minutes.',
  },
  {
    theme: 'brand',
    heading: 'Install labels are available',
    body: 'Label an install to filter it across the dashboard.',
  },
  {
    theme: 'neutral',
    heading: 'No changes to apply',
    body: 'The last sync found no configuration changes.',
  },
]

export const Overview = () => (
  <ComponentDocs
    name="Banner"
    tier="atom"
    summary="A message block that stays in the page, tinted and iconed by severity."
    use={[
      'Keep an important message in place while the user works, such as a failed submission or a plan awaiting review.',
      'Pass actions when the message offers next steps.',
      'Pass a node heading when the message needs a badge or other markup beside its text.',
    ]}
    avoid={[
      'Do not use a banner for transient feedback. That is Toast.',
      'Do not use a banner for a resource state. That is Status.',
      'Do not hand-roll a severity card. Every severity message block is this component.',
      'Do not make a banner dismissible when the user must act on it.',
    ]}
    rules={[
      'A banner and a toast share the status colours and icons, and stay visually distinct: the banner tints its whole surface, the toast keeps a neutral card behind a status rail.',
      'String headings and bodies take the banner typography. Nodes render as given, inheriting caption size so content that sets its own type still wins.',
      'Error and warning announce assertively; every other theme announces politely.',
      'Dismissal collapses the banner, absorbs the surrounding stack gap, then removes it and calls onDismiss once.',
      'Banner persists nothing. A caller that wants a banner to stay dismissed stores that itself in onDismiss.',
      'One primary action at most.',
    ]}
    props={[
      {
        name: 'theme',
        type: 'TBannerTheme',
        default: "'default'",
        description: 'Status colour and default icon.',
      },
      {
        name: 'heading',
        type: 'ReactNode',
        description:
          'Message heading. A string renders at medium body weight; a node renders as given.',
      },
      {
        name: 'icon',
        type: 'TIconVariant',
        description: 'Overrides the theme icon without changing its colour.',
      },
      {
        name: 'actions',
        type: 'ReactNode',
        description: 'Actions, aligned right below the message.',
      },
      {
        name: 'dismissible',
        type: 'boolean',
        default: 'false',
        description: 'Adds an always-visible dismiss button.',
      },
      {
        name: 'dismissLabel',
        type: 'string',
        default: "'Dismiss'",
        description: 'Accessible name for the dismiss button.',
      },
      {
        name: 'onDismiss',
        type: '() => void',
        description: 'Runs once, after the banner has left the layout.',
      },
      {
        name: 'className',
        type: 'string',
        description: 'Additional classes for the collapsing root.',
      },
      {
        name: 'children',
        type: 'ReactNode',
        description:
          'Message body. A string renders as caption text; a node renders as given.',
      },
    ]}
  />
)

export const Themes = () => (
  <div className="flex max-w-3xl flex-col gap-3 p-8">
    {THEMES.map(({ theme, heading, body }) => (
      <Banner key={theme} theme={theme} heading={heading}>
        {body}
      </Banner>
    ))}
  </div>
)

export const Actions = () => (
  <div className="max-w-3xl p-8">
    <Banner
      theme="warn"
      heading="Terraform plan requires review"
      actions={
        <>
          <Button variant="ghost">View plan</Button>
          <Button>Deny</Button>
          <Button variant="primary">Approve</Button>
        </>
      }
    >
      Inspect the proposed infrastructure changes before applying them.
    </Banner>
  </div>
)

const DismissibleDemo = () => {
  const [round, setRound] = useState(0)
  const [dismissals, setDismissals] = useState<string[]>([])

  const recordDismissal = () =>
    setDismissals((current) => [
      ...current,
      `${current.length + 1} · onDismiss at ${DateTime.now().toFormat('HH:mm:ss.SSS')}`,
    ])

  return (
    <div className="flex max-w-3xl flex-col gap-4 p-8">
      <Banner
        key={round}
        theme="info"
        heading="Deploy in progress"
        dismissible
        onDismiss={recordDismissal}
      >
        Deploying payments to production. This may take a few minutes.
      </Banner>
      <div>
        <Button onClick={() => setRound((value) => value + 1)}>
          Reset banner
        </Button>
      </div>
      <div className="flex flex-col gap-1 border-t pt-4">
        <Text variant="label" color="tertiary">
          onDismiss calls
        </Text>
        {dismissals.length === 0 ? (
          <Text variant="caption" color="tertiary">
            Nothing yet. Dismiss the banner, and note that the call lands after
            the banner leaves the layout.
          </Text>
        ) : (
          dismissals.map((dismissal) => (
            <Text
              key={dismissal}
              variant="caption"
              family="mono"
              color="secondary"
            >
              {dismissal}
            </Text>
          ))
        )}
      </div>
    </div>
  )
}

export const Dismissible = () => <DismissibleDemo />

const StackedDemo = () => {
  const [round, setRound] = useState(0)

  return (
    <div className="flex max-w-3xl flex-col gap-6 p-8">
      <div className="flex flex-col gap-3">
        {THEMES.slice(0, 4).map(({ theme, heading }) => (
          <Banner
            key={`${round}-${theme}`}
            theme={theme}
            heading={heading}
            dismissible
          >
            Dismiss this one and the rest close the gap with it.
          </Banner>
        ))}
      </div>
      <div>
        <Button onClick={() => setRound((value) => value + 1)}>
          Reset banners
        </Button>
      </div>
    </div>
  )
}

export const Stacked = () => <StackedDemo />

export const HeadingOnly = () => (
  <div className="max-w-3xl p-8">
    <Banner theme="success" heading="Configuration updated" />
  </div>
)

export const BodyOnly = () => (
  <div className="max-w-3xl p-8">
    <Banner theme="neutral">No configuration changes were found.</Banner>
  </div>
)

export const RichContent = () => (
  <div className="max-w-3xl p-8">
    <Banner
      theme="error"
      heading={
        <div className="flex flex-wrap items-center gap-2">
          <Text weight="medium">Terraform apply failed</Text>
          <Badge variant="code">provider_error</Badge>
        </div>
      }
    >
      <div className="flex flex-col gap-2">
        <Text variant="caption" color="secondary">
          The provider returned an error while creating the resource.
        </Text>
        <Text
          as="pre"
          variant="caption"
          family="mono"
          className="overflow-x-auto rounded-md bg-surface-02 p-2"
        >
          Error: resource &quot;aws_eks_cluster.primary&quot; was not found
        </Text>
      </div>
    </Banner>
  </div>
)

export const IconOverride = () => (
  <div className="max-w-3xl p-8">
    <Banner
      theme="brand"
      icon="GitBranchIcon"
      heading="Branch configuration changed"
    >
      Review the branch before starting a deploy.
    </Banner>
  </div>
)

export const InForm = () => (
  <div className="flex max-w-lg flex-col gap-4 p-8">
    <Banner theme="error" heading="Configuration update failed">
      The configuration changed after this form was opened.
    </Banner>
    <div className="flex flex-col gap-1">
      <Text variant="caption" color="secondary">
        Branch name
      </Text>
      <div className="h-9 rounded-lg border bg-surface-02" />
    </div>
    <div className="flex justify-end gap-2">
      <Button variant="ghost">Cancel</Button>
      <Button variant="primary">Save changes</Button>
    </div>
  </div>
)
