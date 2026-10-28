import { useQuery } from '@tanstack/react-query'
import { DateTime } from 'luxon'
import { useId } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Code } from '@/components/common/Code'
import { Expand } from '@/components/common/Expand'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { useInstallLink } from '@/hooks/use-install-path'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAccount } from '@/lib/ctl-api/accounts'
import { getInstallStack } from '@/lib/ctl-api/installs/get-install-stack'
import { getInstallTelemetryCollector } from '@/lib/ctl-api/installs/install-telemetry-collector'
import { getInstallComponents } from '@/lib/ctl-api/installs/components/get-install-components'
import { listTokens } from '@/lib/ctl-api/tokens/list-tokens'
import type { TInstall } from '@/types'
import { CollectorCredentialModalContainer } from './CollectorCredentialContainer'
import { RemoveCollectorAccessModal } from './RemoveCollectorAccess'

export const CollectorSetupContainer = ({
  install,
  orgId,
  onNavigate,
}: {
  install: TInstall
  orgId: string
  onNavigate: () => void
}) => {
  const id = useId()
  const installId = install.id!
  const installLink = useInstallLink()
  const { addModal } = useSurfaces()
  const account = useQuery({
    queryKey: ['account'],
    queryFn: getAccount,
    staleTime: 60_000,
  })
  const isOrgAdmin =
    !account.isError &&
    !!account.data?.roles?.some(
      (role) => role.org_id === orgId && role.role_type === 'org_admin'
    )
  const collector = useQuery({
    queryKey: ['install-telemetry-collector', orgId, installId],
    queryFn: () => getInstallTelemetryCollector({ orgId, installId }),
    retry: false,
  })
  const components = useQuery({
    queryKey: ['install-components', orgId, installId, 'telemetry-collector'],
    queryFn: () =>
      getInstallComponents({
        orgId,
        installId,
        labels: 'nuon.co/telemetry-collector:true',
        limit: 2,
      }),
    retry: false,
  })
  const tokens = useQuery({
    queryKey: ['tokens', orgId, collector.data?.account_id],
    queryFn: () =>
      listTokens({ orgId, accountId: collector.data?.account_id }),
    enabled: isOrgAdmin && !!collector.data?.account_id,
    retry: false,
  })
  const stack = useQuery({
    queryKey: ['install-stack', orgId, installId],
    queryFn: () => getInstallStack({ orgId, installId }),
    enabled: !!install.runner_id,
    retry: false,
  })
  const declared = components.data?.data ?? []
  const ambiguous = declared.length > 1 || !!components.data?.pagination.hasNext
  const component = !ambiguous ? declared[0] : undefined
  const rawEndpoint =
    stack.data?.install_stack_outputs?.data_contents?.telemetry_endpoint
  const endpoint =
    typeof rawEndpoint === 'string' &&
    /^https?:\/\/\S+$/.test(rawEndpoint.trim())
      ? rawEndpoint.trim()
      : undefined
  const bootstrapTokens = tokens.data ?? []
  const readError =
    collector.isError || components.isError || account.isError || tokens.isError

  return (
    <>
      <section className="flex flex-col gap-4 border-t pt-6">
        <Text weight="strong">Runner collector (default)</Text>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <LabeledValue label="Runner status">
            {install.runner_id ? (
              <Status status={install.runner_status || 'unknown'} />
            ) : (
              'No runner configured'
            )}
          </LabeledValue>
          <LabeledValue
            label="Private telemetry endpoint"
            className="min-w-0"
            loading={!!install.runner_id && stack.isPending}
          >
            {stack.isError ? (
              <div className="flex flex-col gap-2">
                <Text variant="subtext">Unknown — load failed</Text>
                <Button
                  size="sm"
                  variant="ghost"
                  className="w-fit"
                  onClick={() => stack.refetch()}
                >
                  Retry endpoint
                </Button>
              </div>
            ) : endpoint ? (
              <div className="flex min-w-0 flex-col gap-2">
                <Status status="success">Configured</Status>
                <Code
                  variant="inline"
                  className="max-w-full !overflow-x-scroll [&::-webkit-scrollbar]:h-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-cool-grey-400 dark:[&::-webkit-scrollbar-thumb]:bg-cool-grey-600"
                  tabIndex={0}
                  aria-label="Private telemetry endpoint"
                >
                  {endpoint}
                </Code>
              </div>
            ) : (
              'Not configured'
            )}
          </LabeledValue>
        </div>
        <Text variant="subtext">
          The default collection path requires an active runner and private
          telemetry ingress. Saving forwarding settings does not start the
          runner or configure ingress.
        </Text>
      </section>

      <section className="border-t pt-6">
        <Expand
          id={`in-cluster-collector-${installId}`}
          heading={
            <span className="flex flex-col gap-1 text-left">
              <Text weight="strong">In-cluster collector (optional)</Text>
              <Text variant="subtext" theme="neutral">
                Alternative for collection while the runner is stopped.
              </Text>
            </span>
          }
        >
          <div className="flex flex-col gap-4 pt-4">
            <Text variant="subtext">
              Runs as a Nuon component in the cluster. Useful when the runner is
              offline most of the time.
            </Text>
            {readError ? (
              <Banner theme="warn">
                Some setup details could not be loaded. Unknown does not mean
                missing.{' '}
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => {
                    collector.refetch()
                    components.refetch()
                    account.refetch()
                    if (isOrgAdmin && collector.data?.account_id)
                      tokens.refetch()
                  }}
                >
                  Retry setup details
                </Button>
              </Banner>
            ) : null}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <LabeledValue
                label={
                  <Text variant="subtext" theme="neutral">
                    <Tooltip
                      showIcon
                      tabIndex={0}
                      aria-describedby={`${id}-account-help`}
                      className="cursor-help rounded-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-400/80"
                      tipContentClassName="max-w-72 !whitespace-normal"
                      tipContent={
                        <Text variant="subtext" id={`${id}-account-help`}>
                          An identity scoped to this install's telemetry APIs,
                          not a Kubernetes ServiceAccount. Created automatically
                          with the first bootstrap credential.
                        </Text>
                      }
                    >
                      Nuon service account
                    </Tooltip>
                  </Text>
                }
                loading={collector.isPending}
              >
                {collector.isError ? (
                  <Text variant="subtext">Unknown — load failed</Text>
                ) : collector.data?.account_id ? (
                  <Status status="success">Created</Status>
                ) : (
                  <div className="flex flex-col gap-1">
                    <Text variant="subtext">Not created</Text>
                    <Text variant="subtext" theme="neutral">
                      Create a bootstrap credential to set up this account.
                    </Text>
                  </div>
                )}
              </LabeledValue>
              <LabeledValue
                label={
                  <Text variant="subtext" theme="neutral">
                    <Tooltip
                      showIcon
                      tabIndex={0}
                      aria-describedby={`${id}-component-help`}
                      className="cursor-help rounded-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-400/80"
                      tipContentClassName="max-w-72 !whitespace-normal"
                      tipContent={
                        <Text variant="subtext" id={`${id}-component-help`}>
                          The app config declares the collector deployment.
                          Enabled means it is selected for this install, not
                          that it has been deployed.
                        </Text>
                      }
                    >
                      Collector component
                    </Tooltip>
                  </Text>
                }
                loading={components.isPending}
              >
                {components.isError
                  ? 'Unknown — load failed'
                  : ambiguous
                    ? 'Multiple declared — review app config'
                    : component
                      ? component.enabled === false
                        ? 'Disabled in install config'
                        : component.enabled === true
                          ? 'Enabled in install config'
                          : 'Declared in app config'
                      : 'Not declared in app config'}
              </LabeledValue>
              <LabeledValue
                label={
                  <Text variant="subtext" theme="neutral">
                    <Tooltip
                      showIcon
                      tabIndex={0}
                      aria-describedby={`${id}-cluster-help`}
                      className="cursor-help rounded-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-400/80"
                      tipContentClassName="max-w-72 !whitespace-normal"
                      tipContent={
                        <Text variant="subtext" id={`${id}-cluster-help`}>
                          The collector needs a credential Secret, queue
                          storage, and network access. Nuon does not check these
                          here; check the component's deployment details.
                        </Text>
                      }
                    >
                      Cluster prerequisites
                    </Tooltip>
                  </Text>
                }
              >
                Not verified
              </LabeledValue>
            </div>
            {component?.component_id ? (
              <Link
                href={installLink({
                  orgId,
                  installId,
                  appId: install.app_id,
                  suffix: `/components/${component.component_id}`,
                })}
                onClick={onNavigate}
              >
                View collector component
              </Link>
            ) : null}

            <Text weight="strong">Bootstrap credentials</Text>
            {account.isPending ? (
              <Text loading>Checking permissions</Text>
            ) : !isOrgAdmin ? (
              <Text variant="subtext">
                {account.isError
                  ? 'Permission check failed. Retry setup details before managing credentials.'
                  : 'An org admin must create, rotate, or revoke collector credentials.'}
              </Text>
            ) : (
              <>
                <div className="flex flex-wrap gap-3">
                  <Button
                    variant="secondary"
                    disabled={collector.isPending || collector.isError}
                    tooltipProps={
                      collector.isPending || collector.isError
                        ? { tipContent: 'Load collector account details first' }
                        : undefined
                    }
                    onClick={() =>
                      addModal(
                        <CollectorCredentialModalContainer
                          orgId={orgId}
                          installId={installId}
                        />
                      )
                    }
                  >
                    Create bootstrap credential
                  </Button>
                  {collector.data?.account_id && !collector.isError ? (
                    <Button
                      variant="ghost"
                      onClick={() =>
                        addModal(
                          <RemoveCollectorAccessModal
                            orgId={orgId}
                            installId={installId}
                          />
                        )
                      }
                    >
                      Remove collector access
                    </Button>
                  ) : null}
                </div>
                {collector.data?.account_id ? (
                  tokens.isError ? (
                    <Text variant="subtext" theme="warn">
                      Credential metadata is unknown — load failed.
                    </Text>
                  ) : (
                    <Table
                      data={bootstrapTokens}
                      isLoading={tokens.isPending}
                      skeletonRows={1}
                      enableSearch={false}
                      enableSorting={false}
                      emptyStateProps={{
                        emptyTitle: 'No bootstrap credentials issued',
                        emptyMessage:
                          'Create a credential for the sensitive collector input.',
                      }}
                      columns={[
                        {
                          header: 'Credential',
                          accessorKey: 'id',
                          cell: ({ row }) => (
                            <Text family="mono" variant="subtext">
                              {row.original.id}
                            </Text>
                          ),
                        },
                        {
                          header: 'Expiry',
                          accessorKey: 'expires_at',
                          cell: ({ row }) =>
                            row.original.expires_at ? (
                              <div className="flex flex-col gap-1">
                                <Time time={row.original.expires_at} />
                                <Text
                                  variant="subtext"
                                  theme={
                                    DateTime.fromISO(row.original.expires_at) <=
                                    DateTime.now()
                                      ? 'warn'
                                      : 'neutral'
                                  }
                                >
                                  {DateTime.fromISO(row.original.expires_at) <=
                                  DateTime.now()
                                    ? 'Expired'
                                    : 'Issued — use not verified'}
                                </Text>
                              </div>
                            ) : (
                              <Text variant="subtext">Unknown</Text>
                            ),
                        },
                        {
                          header: 'Action',
                          id: 'action',
                          cell: ({ row }) => (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() =>
                                addModal(
                                  <RemoveCollectorAccessModal
                                    orgId={orgId}
                                    installId={installId}
                                    token={row.original}
                                  />
                                )
                              }
                            >
                              Revoke credential
                            </Button>
                          ),
                        },
                      ]}
                    />
                  )
                ) : null}
              </>
            )}
            <Expand
              id={`collector-setup-${installId}`}
              heading="Set up the in-cluster collector"
            >
              <div className="flex flex-col gap-3 pt-3">
                <ol className="list-decimal pl-5 space-y-3 [&>li]:list-item">
                  <Text as="li" variant="subtext">
                    <Text variant="subtext" weight="strong">
                      Add the Helm component.
                    </Text>{' '}
                    Use Nuon's telemetry-agent chart as an optional app
                    component. Configure its namespace, install identity, image,
                    queue storage, and network access.
                    <Link
                      isExternal
                      href="https://github.com/nuonco/charts/tree/main/charts/telemetry-agent"
                      className="mt-1"
                    >
                      View Nuon telemetry-agent chart
                    </Link>
                  </Text>
                  <Text as="li" variant="subtext">
                    <Text variant="subtext" weight="strong">
                      Label the collector component.
                    </Text>{' '}
                    Add this to its component config. Exactly one component
                    should have this label:
                    <Code
                      variant="preformated"
                      className="min-h-0 max-w-full p-2 mt-2"
                    >
                      {'[labels]\n"nuon.co/telemetry-collector" = "true"'}
                    </Code>
                  </Text>
                  <Text as="li" variant="subtext">
                    <Text variant="subtext" weight="strong">
                      Add the sensitive input.
                    </Text>{' '}
                    Define an optional string install input with{' '}
                    <Code variant="inline">sensitive = true</Code> and pass its
                    value to the chart's{' '}
                    <Code variant="inline">bootstrapToken</Code> value. The
                    chart creates and mounts the Kubernetes Secret.
                  </Text>
                  <Text as="li" variant="subtext">
                    <Text variant="subtext" weight="strong">
                      Sync the app config.
                    </Text>{' '}
                    Run <Code variant="inline">nuon apps sync</Code> from the
                    app config directory to publish the component, label, and
                    input. This does not deploy existing installs.
                  </Text>
                  <Text as="li" variant="subtext">
                    <Text variant="subtext" weight="strong">
                      Create and save the credential.
                    </Text>{' '}
                    As an org admin, select Create bootstrap credential and copy
                    the one-time token into this install's sensitive collector
                    input. For config-managed installs, sync a protected local
                    install config. Never commit the token.
                  </Text>
                  <Text as="li" variant="subtext">
                    <Text variant="subtext" weight="strong">
                      Deploy and verify.
                    </Text>{' '}
                    With the runner online, apply the synced app config and
                    enable/deploy the component through its normal workflow.
                    Route application OTLP exporters to its cluster-local
                    Service and enable telemetry forwarding. Check fresh data in
                    the destination before stopping the runner. Avoid sending
                    the same signals through both collectors.
                  </Text>
                </ol>
                <Text variant="subtext" weight="strong">
                  Rotate a credential
                </Text>
                <ol className="list-decimal pl-5 space-y-2 [&>li]:list-item">
                  <Text as="li" variant="subtext">
                    Before expiry, create a replacement bootstrap credential.
                    Existing credentials remain valid.
                  </Text>
                  <Text as="li" variant="subtext">
                    Bring the runner online, update the sensitive input, and
                    deploy the component change.
                  </Text>
                  <Text as="li" variant="subtext">
                    Verify the collector uses the replacement, then select
                    Revoke credential for the old token.
                  </Text>
                </ol>
              </div>
            </Expand>
          </div>
        </Expand>
      </section>
    </>
  )
}
