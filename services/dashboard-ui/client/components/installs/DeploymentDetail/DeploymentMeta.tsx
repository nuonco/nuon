import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { Card } from '@/components/common/Card'
import { CommitLink } from '@/components/common/GitReferenceLink'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { getAppBranch, getBranchWorkflowRun, getInstallUpdates } from '@/lib'
import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { humanize } from '@/utils/string-utils'
import { vcsRepo } from '@/utils/vcs-urls'
import { isBranchRunDeployment } from './deployment-progress'
import {
  branchFromAppBranchRun,
  deploymentRunBranch,
  type TDeploymentRunBranch,
} from './deployment-run-branch'

export type TDeploymentBranchMeta = {
  name: string
  href?: string
  runHref?: string
  sha?: string
  message?: string
  author?: string
  avatarUrl?: string
  createdAt?: string
  repo?: string
  isLoading?: boolean
}

export const useDeploymentRunBranch = ({
  orgId,
  appId,
  type,
  installId,
  deployment,
  workflow,
  repo,
}: {
  orgId?: string
  appId?: string
  type?: string
  installId?: string
  deployment?: Pick<TInstallDeploymentRecord, 'app_branch'> | null
  workflow?: TWorkflow | null
  repo?: string
}): TDeploymentBranchMeta | undefined => {
  const ref = deployment?.app_branch
  const needsRun = !workflow?.app_branch_runs?.length && !!ref?.run_id
  const { data: branchWorkflow, isLoading: runLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-workflow-run', orgId, appId, ref?.id, ref?.run_id],
    queryFn: () =>
      getBranchWorkflowRun({
        orgId: orgId!,
        appId: appId!,
        branchId: ref!.id,
        runId: ref!.run_id!,
      }),
    enabled: needsRun && !!orgId && !!appId,
  })
  const own = deploymentRunBranch(
    deployment,
    branchWorkflow?.app_branch_runs?.length ? branchWorkflow : workflow
  )
  const needsApplied =
    !isBranchRunDeployment(type) && !needsRun && !own?.sha && !own?.message
  const { data: updates, isLoading: appliedLoading } = useQuery({
    queryKey: ['install-current-app-branch-run', orgId, installId],
    queryFn: () =>
      getInstallUpdates({
        orgId: orgId!,
        installId: installId!,
        page: 0,
        limit: 1,
      }),
    enabled: needsApplied && !!orgId && !!installId,
  })
  const resolved: TDeploymentRunBranch | undefined =
    own?.sha || own?.message
      ? own
      : (branchFromAppBranchRun(updates?.current_app_branch_run) ?? own)
  const needsBranch = !!resolved?.id && (!resolved.name || !resolved.repo)
  const { data: branch } = useQuery({
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
  const name = resolved?.name || branch?.name
  const loading = (needsRun && runLoading) || (needsApplied && appliedLoading)
  if ((!resolved || !name) && loading)
    return { name: name || '', isLoading: true }
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
    sha: resolved.sha,
    message: resolved.message,
    author: resolved.author,
    avatarUrl: resolved.avatarUrl,
    createdAt: resolved.createdAt,
    repo:
      resolved.repo ||
      (branch ? vcsRepo(latestBranchConfig(branch)) : undefined) ||
      repo,
    isLoading: loading,
  }
}

const BranchCommit = ({ branch }: { branch: TDeploymentBranchMeta }) =>
  branch.message ? (
    <BranchRunCommit
      displayVariant="inline"
      href={branch.runHref}
      message={branch.message}
      author={branch.author}
      avatarUrl={branch.avatarUrl}
      sha={branch.sha}
      repo={branch.repo}
      createdAt={branch.createdAt}
      showStatus={false}
    />
  ) : (
    <CommitLink sha={branch.sha} repo={branch.repo} showIcon={false} />
  )

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
      {branch && (branch.name || branch.isLoading) ? (
        <LabeledValue
          label="App branch"
          loading={branch.isLoading && !branch.name}
          loadingWidth={12}
        >
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
      {branch && (branch.isLoading || branch.sha || branch.message) ? (
        <LabeledValue
          label="Commit"
          className="!shrink grow basis-64 min-w-0 max-w-full"
          loading={branch.isLoading}
          loadingWidth={48}
        >
          <BranchCommit branch={branch} />
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
