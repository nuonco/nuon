import type { ReactNode } from 'react'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'

export interface IBranchHeaderMeta {
  configuration: ReactNode
  repo?: string
  gitBranch?: string
  directory?: string
  trigger?: ReactNode
}

const repoHref = (repo: string) =>
  repo.startsWith('http') ? repo : `https://github.com/${repo}`

const Separator = () => (
  <Text as="span" variant="subtext" theme="neutral" aria-hidden>
    ·
  </Text>
)

export const BranchHeaderMeta = ({
  configuration,
  repo,
  gitBranch,
  directory,
  trigger,
}: IBranchHeaderMeta) => {
  const directoryPath =
    directory && directory !== '.' && directory !== '/' ? directory : undefined
  const hasTracking = !!(repo || gitBranch || directoryPath)

  return (
    <div className="flex flex-wrap items-start gap-x-8 gap-y-3">
      <LabeledValue label="Configuration">{configuration}</LabeledValue>
      {hasTracking ? (
        <LabeledValue label="Tracking" className="md:border-l md:pl-8">
          <span className="flex flex-wrap items-center gap-x-1.5 gap-y-1">
            {repo ? (
              <Link href={repoHref(repo)} isExternal>
                {repo}
              </Link>
            ) : null}
            {gitBranch ? (
              <>
                {repo ? <Separator /> : null}
                <Text variant="subtext" family="mono" flex>
                  <Icon variant="GitBranchIcon" />
                  {gitBranch}
                </Text>
              </>
            ) : null}
            {directoryPath ? (
              <>
                {repo || gitBranch ? <Separator /> : null}
                <Text variant="subtext" family="mono" theme="neutral">
                  {directoryPath}
                </Text>
              </>
            ) : null}
          </span>
        </LabeledValue>
      ) : null}
      {trigger ? (
        <LabeledValue label="Runs on" className="md:border-l md:pl-8">
          <Text variant="subtext" flex>
            <Icon variant="LightningIcon" />
            {trigger}
          </Text>
        </LabeledValue>
      ) : null}
    </div>
  )
}
