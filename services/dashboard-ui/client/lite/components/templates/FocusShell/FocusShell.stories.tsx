import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Card } from '../../atoms/Card'
import { Link } from '../../atoms/Link'
import { Text } from '../../atoms/Text'
import { UserDropdown } from '../../organisms/UserDropdown'
import { FocusShell } from './FocusShell'

export default {
  title: 'lite/templates/FocusShell',
}

const Actions = () => (
  <>
    <Link href="https://docs.nuon.co" external variant="caption">
      Developer docs
    </Link>
    <UserDropdown
      user={{ name: 'Alex Morgan', email: 'alex@example.com' }}
      signOutHref="https://auth.example.com/logout"
    />
  </>
)

const FocusedContent = ({ sections = 3 }: { sections?: number }) => (
  <div className="mx-auto flex w-full max-w-3xl flex-col gap-6">
    <div className="flex flex-col gap-2">
      <Text as="h1" variant="title">
        Configure your workspace
      </Text>
      <Text as="p" variant="caption" color="secondary">
        Complete the required settings before continuing.
      </Text>
    </div>
    {Array.from({ length: sections }, (_, index) => (
      <Card key={index} className="min-h-36">
        <Text weight="semibold">Configuration section {index + 1}</Text>
        <Text as="p" variant="caption" color="secondary">
          Focused-flow content remains independent from the application frame.
        </Text>
      </Card>
    ))}
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="FocusShell"
    tier="template"
    summary="The sidebar-free application frame for onboarding and focused flows."
    use={[
      'Use for top-level routes that need Nuon chrome without dashboard navigation.',
      'Place wizard and focused-flow UI inside the centered content frame.',
    ]}
    avoid={[
      'Do not add dashboard navigation, organization status, or a status bar.',
      'Do not put wizard progress, skip, or exit behavior in the shell.',
    ]}
    rules={[
      'The header sticks inside the main scrollbar and gains glass elevation after scrolling.',
      'Default content is guttered and capped at 72rem.',
      'Use fullBleed only for content that genuinely owns edge-to-edge layout.',
    ]}
    props={[
      {
        name: 'actions',
        type: 'ReactNode',
        description: 'Global controls passed to FocusHeader.',
      },
      {
        name: 'fullBleed',
        type: 'boolean',
        default: 'false',
        description: 'Removes the default maximum width and content gutters.',
      },
      {
        name: 'contentClassName',
        type: 'string',
        description: 'Additional constraints for the content frame.',
      },
    ]}
  />
)

export const Default = () => (
  <FocusShell actions={<Actions />}>
    <FocusedContent />
  </FocusShell>
)
Default.meta = { fullBleed: true }

export const Scrolling = () => (
  <FocusShell actions={<Actions />}>
    <FocusedContent sections={10} />
  </FocusShell>
)
Scrolling.meta = { fullBleed: true }

export const NarrowContent = () => (
  <FocusShell actions={<Actions />} contentClassName="max-w-4xl">
    <FocusedContent />
  </FocusShell>
)
NarrowContent.meta = { fullBleed: true }

export const EmptyActions = () => (
  <FocusShell>
    <FocusedContent />
  </FocusShell>
)
EmptyActions.meta = { fullBleed: true }

export const FullBleed = () => (
  <FocusShell actions={<Actions />} fullBleed>
    <div className="grid min-h-full grid-cols-1 bg-surface-01 md:grid-cols-2">
      <div className="flex items-center justify-center p-8">
        <Text variant="heading">Focused content</Text>
      </div>
      <div className="flex items-center justify-center bg-surface-02 p-8">
        <Text variant="heading">Supporting context</Text>
      </div>
    </div>
  </FocusShell>
)
FullBleed.meta = { fullBleed: true }
