import { useQuery } from '@tanstack/react-query'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { Card } from '@/components/common/Card'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { getAppBranch } from '@/lib'
import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { humanize } from '@/utils/string-utils'
import { vcsRepo } from '@/utils/vcs-urls'
import { isBranchRunDeployment } from './deployment-progress'
import { deploymentRunBranch } from './deployment-run-branch'

export type TDeploymentBranchMeta = {
  name: string
  href?: string
  runHref?: string
  status?: string
  sha?: string
  message?: string
  author?: string
  avatarUrl?: string
  createdAt?: string
  repo?: string
}

export const useDeploymentRunBranch = ({
  orgId,
  appId,
  type,
  deployment,
  workflow,
  repo,
}: {
  orgId?: string
  appId?: string
  type?: string
  deployment?: Pick<TInstallDeploymentRecord, 'app_branch'> | null
  workflow?: TWorkflow | null
  repo?: string
}): TDeploymentBranchMeta | undefined => {
  const isBranchRun = isBranchRunDeployment(type)
  const resolved = isBranchRun
    ? deploymentRunBranch(deployment, workflow)
    : undefined
  const needsBranch = !!resolved?.id && (!resolved.name || !resolved.repo)
  const { data } = useQuery({
    queryKey: ['deployment-run-branch', orgId, appId, resolved?.id],
    queryFn: () =>
      getAppBranch({
        orgId: orgId!,
        appId: appId!,
        branchId: resolved!.id,
        latestConfig: true,
      }),
    enabled: needsBranch && !!orgId && !!appId,
  })
  const name = resolved?.name || data?.name
  if (!resolved || !name) return undefined
  const href =
    orgId && appId
      ? `/${orgId}/apps/${appId}/branches/${resolved.id}`
      : undefined
  return {
    name,
    href,
    runHref:
      href && resolved.runId ? `${href}/runs/${resolved.runId}` : undefined,
    status: resolved.status,
    sha: resolved.sha,
    message: resolved.message,
    author: resolved.author,
    avatarUrl: resolved.avatarUrl,
    createdAt: resolved.createdAt,
    repo:
      resolved.repo ||
      (data ? vcsRepo(latestBranchConfig(data)) : undefined) ||
      repo,
  }
}

export const DeploymentMeta = ({
  type,
  branch,
  framed = false,
}: {
  type?: string
  branch?: TDeploymentBranchMeta
  framed?: boolean
}) => {
  if (!type && !branch) return null
  const fields = (
    <>
      {type ? <LabeledValue label="Type">{humanize(type)}</LabeledValue> : null}
      {branch ? (
        <LabeledValue label="App branch">
          {branch.href ? (
            <Link href={branch.href} textVariant="subtext">
              <Text as="span" variant="subtext" family="mono">
                {branch.name}
              </Text>
            </Link>
          ) : (
            <Text as="span" variant="subtext" family="mono">
              {branch.name}
            </Text>
          )}
        </LabeledValue>
      ) : null}
      {branch?.sha || branch?.message ? (
        <LabeledValue
          label="Commit"
          className="!shrink grow basis-64 min-w-0 max-w-full"
        >
          <BranchRunCommit
            displayVariant="inline"
            status={branch.status}
            href={branch.runHref}
            message={branch.message}
            author={branch.author}
            avatarUrl={branch.avatarUrl}
            sha={branch.sha}
            repo={branch.repo}
            createdAt={branch.createdAt}
            showStatus={false}
          />
        </LabeledValue>
      ) : null}
    </>
  )
  if (!framed) return fields
  return (
    <Card className="!p-4 !shadow-none w-full">
      <div className="flex flex-wrap items-start gap-x-8 gap-y-4 [&>*]:shrink-0 [&>*]:whitespace-nowrap">
        {fields}
      </div>
    </Card>
  )
}
