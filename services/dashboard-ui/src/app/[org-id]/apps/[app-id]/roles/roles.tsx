import { Link } from '@/components/common/Link'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { CodeBlock } from '@/components/common/CodeBlock'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { getAppConfigs, getAppConfigById } from '@/lib'
import { decodeAsString } from "@/utils/data-utils"

export const AppRoles = async ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) => {
  const { data: configs, error: configsError } = await getAppConfigs({
    appId,
    orgId,
  })

  if (configsError) {
    return <AppRolesError />
  }

  const { data: config, error } = await getAppConfigById({
    appConfigId: configs?.at(0)?.id,
    appId,
    orgId,
    recurse: true,
  })

  return error ? (
    <AppRolesError />
  ) : config?.permissions?.aws_iam_roles?.length ? (
    <div className="flex flex-col divide-y">
      {config?.permissions?.aws_iam_roles?.map((role) => (
        <div className="flex flex-col gap-4 py-8" key={role?.id}>
          <div className="flex flex-col">
            <Text variant="h3" weight="strong">
              {role?.display_name}
            </Text>
            <Text variant="subtext" theme="neutral">
              {role?.description}
            </Text>
          </div>

          {role?.permissions_boundary ? (
            <div className="flex flex-col gap-2">
              <Text variant="subtext" weight="strong">
                Permission boundary
              </Text>
              <CodeBlock language="json">
                {decodeAsString(role?.permissions_boundary)}
              </CodeBlock>
            </div>
          ) : null}

          <Card>
            <Text weight="strong">Policies</Text>
            {role?.policies?.map((policy) => (
              <div key={policy?.id} className="flex flex-col gap-2 my-2">
                {policy?.managed_policy_name ? (
                  <>
                    <Text variant="subtext" weight="strong">
                      AWS managed policy
                    </Text>
                    <Code variant="inline" className="!px-2">
                      <Link
                        href={`https://docs.aws.amazon.com/aws-managed-policy/latest/reference/${policy?.managed_policy_name}.html`}
                        isExternal
                      >
                        <Text family="mono">{policy?.managed_policy_name}</Text>
                        <Icon variant="ArrowSquareOut" />
                      </Link>
                    </Code>
                  </>
                ) : null}
                {policy?.contents ? (
                  <>
                    <Text variant="subtext" weight="strong">
                      {policy?.name}
                    </Text>
                    <CodeBlock language="json">
                      {decodeAsString(policy?.contents)}
                    </CodeBlock>
                  </>
                ) : null}
              </div>
            ))}
          </Card>
        </div>
      ))}
    </div>
  ) : (
    <AppRolesError
      {...{
        title: 'No roles found',
        message:
          "You don't have any roles assigned yet. Contact your administrator to get access to roles.",
      }}
    />
  )
}

export const AppRolesSkeleton = () => {
  return (
    <div className="flex flex-col divide-y">
      {Array.from({ length: 4 }).map((_, idx) => (
        <div className="flex flex-col gap-4 py-8" key={idx}>
          <div className="flex flex-col gap-1">
            <Skeleton width="250px" height="27px" />
            <Skeleton width="300px" height="17px" />
          </div>

          <div className="flex flex-col gap-2">
            <Skeleton height="17px" width="118px" />
            <Skeleton height="80px" width="100%" />
          </div>

          <Card>
            <Skeleton height="24px" width="50px" />
            <div className="flex flex-col gap-2 my-2">
              <Skeleton height="17px" width="118px" />
              <Skeleton height="32px" width="200px" />
            </div>

            <div className="flex flex-col gap-2 my-2">
              <Skeleton height="17px" width="250px" />
              <Skeleton height="300px" width="100%" />
            </div>
          </Card>
        </div>
      ))}
    </div>
  )
}

export const AppRolesError = ({
  title = 'Unable to load roles',
  message = 'We encountered an issue loading your roles. Please try refreshing the page or contact support if the problem persists.',
}: {
  title?: string
  message?: string
}) => {
  return (
    <EmptyState variant="table" emptyMessage={message} emptyTitle={title} />
  )
}
