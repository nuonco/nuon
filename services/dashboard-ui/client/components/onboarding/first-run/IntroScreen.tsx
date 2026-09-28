import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Logo } from '@/components/common/Logo'
import { Text } from '@/components/common/Text'
import { DOCS_URL } from './constants'
import { MiniArch, NuonMark } from './shared'

const IntroDiagram = () => (
  <div className="flex flex-col gap-3">
    <div className="flex flex-col gap-3 rounded-lg border bg-background p-4 shadow-sm">
      <div className="flex items-center gap-2">
        <Icon variant="GitBranchIcon" size={18} theme="neutral" />
        <Text variant="base" weight="strong">
          Your app template
        </Text>
      </div>
      <MiniArch />
    </div>

    <div className="flex items-center justify-center gap-3 py-1">
      <NuonMark className="h-7 w-auto text-neutral-900 dark:text-white" />
      <Icon variant="ArrowDownIcon" size={24} weight="bold" theme="neutral" />
    </div>

    {/* Offset rings behind the account stand in for "every customer". */}
    <div className="relative">
      <div
        className="absolute inset-0 translate-x-3 translate-y-3 rounded-xl ring-2 ring-primary-200 dark:ring-primary-900"
        aria-hidden
      />
      <div
        className="absolute inset-0 translate-x-1.5 translate-y-1.5 rounded-xl ring-2 ring-primary-300 dark:ring-primary-800"
        aria-hidden
      />
      <div className="relative flex flex-col gap-3 rounded-xl p-4 bg-primary-50 dark:bg-primary-950/60 ring-2 ring-primary-500 shadow-md">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Icon variant="CloudIcon" size={20} weight="fill" theme="brand" />
            <Text variant="body" weight="strong">
              Your customer&apos;s cloud account
            </Text>
          </div>
          <div className="flex items-center gap-2">
            <Icon variant="AWSColor" size={18} />
            <Icon variant="GCPColor" size={16} />
            <Icon variant="AzureColor" size={16} />
          </div>
        </div>
        <div className="flex flex-col gap-3 rounded-lg border bg-background p-4 shadow-sm">
          <div className="flex items-center justify-between gap-3">
            <Text variant="base" weight="strong">
              Your app
            </Text>
            <div className="flex items-center gap-1.5">
              <span className="animate-pulse">
                <Icon variant="CheckCircleIcon" size={16} weight="fill" theme="success" />
              </span>
              <Text variant="subtext" weight="strong" theme="success">
                Running
              </Text>
            </div>
          </div>
          <MiniArch live />
        </div>
      </div>
    </div>
  </div>
)

// Sits before the stepper and reads as an extension of sign-in: one sentence,
// one button, one diagram. The org already exists by the time it renders.
export const IntroScreen = ({ onStart }: { onStart: () => void }) => (
  <div className="h-screen flex flex-col bg-background overflow-y-auto">
    <div className="flex justify-between w-full px-6 pt-4">
      <Logo />
      <Button variant="ghost" href={DOCS_URL} size="sm">
        <Icon variant="BookOpenIcon" size={14} /> Docs
      </Button>
    </div>
    <div className="flex-1 flex px-6 py-12">
      {/* my-auto centers when there is room and top-aligns when the content is taller than the viewport. */}
      <div className="max-w-5xl mx-auto my-auto w-full grid gap-10 md:grid-cols-[1fr_1.2fr] items-center">
        <div className="flex flex-col gap-8">
          <div className="flex flex-col gap-3">
            <Text variant="h1" role="heading" level={1}>
              Your account is set up
            </Text>
            <Text variant="base" theme="neutral">
              The first step is to create an app template that you can deploy to a customer&apos;s
              cloud.
            </Text>
          </div>
          <div>
            <Button variant="primary" size="lg" onClick={onStart}>
              Create your first app template <Icon variant="CaretRightIcon" weight="bold" />
            </Button>
          </div>
        </div>
        <IntroDiagram />
      </div>
    </div>
  </div>
)
