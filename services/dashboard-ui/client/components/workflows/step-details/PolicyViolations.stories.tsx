export default {
  title: 'Workflows/PolicyViolations',
}

import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { PolicyViolations } from './PolicyViolations'
import type { TWorkflowStep } from '@/types'

const violationStep = {
  id: 'step-1',
  policy_results: {
    results: [
      {
        action: 'warn',
        message:
          "EKS cluster 'module.eks.aws_eks_cluster.this[0]' allows public API access from 0.0.0.0/0 — restrict public_access_cidrs to known operator/runner IP ranges",
        policy_id: 'pol-1',
      },
      {
        action: 'warn',
        message:
          "EKS cluster 'module.eks.aws_eks_cluster.this[0]' has public endpoint access enabled — ensure this is intentional e.g., an example app or for demonstrating policies in Nuon",
        policy_id: 'pol-2',
      },
    ],
  },
} as TWorkflowStep

const passedStep = {
  id: 'step-2',
  policy_results: {
    results: [{ action: 'allow' }],
  },
} as TWorkflowStep

export const WithViolations = () => <PolicyViolations step={violationStep} />

export const AllPassed = () => <PolicyViolations step={passedStep} />

// ---------------------------------------------------------------------------
// Layout exploration: each story below renders the same two warnings using a
// different presentation. Pick the one that feels best.
// ---------------------------------------------------------------------------

const messages = [
  "EKS cluster 'module.eks.aws_eks_cluster.this[0]' allows public API access from 0.0.0.0/0 — restrict public_access_cidrs to known operator/runner IP ranges",
  "EKS cluster 'module.eks.aws_eks_cluster.this[0]' has public endpoint access enabled — ensure this is intentional e.g., an example app or for demonstrating policies in Nuon",
]

const splitOnEmDash = (m: string): { issue: string; fix?: string } => {
  const idx = m.indexOf(' — ')
  if (idx === -1) return { issue: m }
  return { issue: m.slice(0, idx), fix: m.slice(idx + 3) }
}

const splitOnQuote = (
  m: string
): { before: string; resource?: string; after: string } => {
  const match = m.match(/^(.*?)'([^']+)'(.*)$/)
  if (!match) return { before: m, after: '' }
  return { before: match[1], resource: match[2], after: match[3] }
}

const Frame = ({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) => (
  <div className="flex flex-col gap-2 max-w-3xl">
    <Text variant="subtext" theme="neutral" weight="strong">
      {title}
    </Text>
    <Card className="!p-0 overflow-hidden">
      <div className="flex flex-col">
        <div className="flex items-center gap-2 p-3 border-b border-cool-grey-200 dark:border-dark-grey-600 text-orange-600 dark:text-orange-500">
          <Icon variant="WarningIcon" size={14} />
          <Text variant="subtext" weight="strong">
            Policy Warnings
          </Text>
        </div>
        {children}
      </div>
    </Card>
  </div>
)

// 1) Two-line: split issue and remediation
export const Variant1_TwoLine = () => (
  <Frame title="1) Two-line: issue on top, remediation below">
    <ul className="px-4 py-3 space-y-3">
      {messages.map((m, i) => {
        const { issue, fix } = splitOnEmDash(m)
        return (
          <li key={i} className="flex flex-col gap-0.5">
            <Text variant="subtext" weight="strong">
              {issue}
            </Text>
            {fix ? (
              <Text variant="subtext" theme="neutral">
                {fix}
              </Text>
            ) : null}
          </li>
        )
      })}
    </ul>
  </Frame>
)

// 2) Inline code chip for quoted resource
export const Variant2_CodeChip = () => (
  <Frame title="2) Inline code chip for the resource">
    <ul className="px-4 py-3 space-y-2">
      {messages.map((m, i) => {
        const { before, resource, after } = splitOnQuote(m)
        return (
          <li key={i} className="flex items-start gap-2">
            <Icon
              variant="CaretRightIcon"
              size={12}
              className="mt-1 shrink-0"
            />
            <Text variant="subtext">
              {before}
              {resource ? (
                <code className="px-1 py-0.5 mx-0.5 rounded bg-cool-grey-100 dark:bg-dark-grey-700 font-mono text-xs">
                  {resource}
                </code>
              ) : null}
              {after}
            </Text>
          </li>
        )
      })}
    </ul>
  </Frame>
)

// 3) Numbered list with left-border accent
export const Variant3_NumberedAccent = () => (
  <Frame title="3) Numbered list with left-border accent">
    <ol className="px-4 py-3 space-y-3">
      {messages.map((m, i) => {
        const { issue, fix } = splitOnEmDash(m)
        return (
          <li
            key={i}
            className="flex gap-3 pl-3 border-l-2 border-orange-400 dark:border-orange-500/60"
          >
            <Text
              variant="subtext"
              weight="strong"
              className="shrink-0 text-orange-600 dark:text-orange-500"
            >
              {i + 1}.
            </Text>
            <div className="flex flex-col gap-0.5 min-w-0">
              <Text variant="subtext" weight="strong">
                {issue}
              </Text>
              {fix ? (
                <Text variant="subtext" theme="neutral">
                  {fix}
                </Text>
              ) : null}
            </div>
          </li>
        )
      })}
    </ol>
  </Frame>
)

// 4) Subtle bordered card per row with hover
export const Variant4_RowCards = () => (
  <Frame title="4) Bordered card per row with hover">
    <div className="flex flex-col gap-2 px-4 py-3">
      {messages.map((m, i) => {
        const { issue, fix } = splitOnEmDash(m)
        return (
          <div
            key={i}
            className={cn(
              'flex items-start gap-3 p-3 rounded-md border',
              'border-cool-grey-200 dark:border-dark-grey-600',
              'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-700/40 transition-colors'
            )}
          >
            <Icon
              variant="WarningIcon"
              size={14}
              className="mt-1 shrink-0 text-orange-600 dark:text-orange-500"
            />
            <div className="flex flex-col gap-0.5 min-w-0">
              <Text variant="subtext" weight="strong">
                {issue}
              </Text>
              {fix ? (
                <Text variant="subtext" theme="neutral">
                  {fix}
                </Text>
              ) : null}
            </div>
          </div>
        )
      })}
    </div>
  </Frame>
)
