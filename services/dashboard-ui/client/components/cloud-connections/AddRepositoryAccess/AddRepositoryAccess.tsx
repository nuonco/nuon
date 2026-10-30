import { useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TAPIError, TCloudConnection } from '@/types'
import { RepositoryPolicyFields } from '../RepositoryPolicyFields'
import {
  parseRepositoryNames,
  repositoryNamesSchema,
} from '../repository-scope'

export const AddRepositoryAccess = ({
  connection,
  onGenerate,
  isPending = false,
  error,
  result,
  repositories = '',
}: {
  connection: TCloudConnection
  onGenerate: (names: string[]) => void
  isPending?: boolean
  error?: TAPIError | null
  result?: {
    names: string[]
    permissionsCli: string
    permissionsPolicy: Record<string, unknown>
  }
  repositories?: string
}) => {
  const { preset, target_id: target, default_region: region } = connection
  const missingRegion = preset !== 'acr' && !region
  const [showGenerator, setShowGenerator] = useState(
    !!repositories || !!result || isPending || !!error
  )
  const schema = z.object({ repositories: repositoryNamesSchema(preset) })
  const form = useForm({
    defaultValues: { repositories },
    validators: { onMount: schema, onChange: schema },
    onSubmit: ({ value }) => {
      if (!missingRegion && !isPending)
        onGenerate(parseRepositoryNames(value.repositories))
    },
  })
  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const value = useStore(form.store, (s) => s.values.repositories)
  const currentResult =
    result && result.names.join('\n') === parseRepositoryNames(value).join('\n')
      ? result
      : undefined
  const name = currentResult?.names[0]
  const example =
    preset === 'ecr'
      ? `[aws_ecr]\nregion = "${region}"\nimage_url = "${target}.dkr.ecr.${region}.amazonaws.com/${name}"`
      : preset === 'gar'
        ? `[gcp_gar]\ngcp_project_id = "${target}"\nregion = "${region}"\nimage_url = "${region}-docker.pkg.dev/${target}/${name}/example-image"`
        : `[azure_acr]\nregistry_url = "${connection.azure_registry_login_server}"\nimage_url = "${connection.azure_registry_login_server}/${name}"`
  return (
    <section
      id="runbook-step-3"
      className="flex min-w-0 flex-col gap-4 border-t pt-6 scroll-mt-6"
      aria-label="Repository access"
    >
      <SectionHeader
        title="Repository access"
        description="Optionally generate commands to grant this identity access to image repositories."
      />
      {!showGenerator ? (
        <Button
          className="w-fit"
          onClick={() => {
            setShowGenerator(true)
            requestAnimationFrame(() =>
              document
                .getElementById(`repository-access-${connection.id}-0`)
                ?.focus()
            )
          }}
        >
          Generate repository access commands
        </Button>
      ) : (
        <>
          <Text variant="subtext" theme="neutral">
            Repository names are used only to generate commands. They are not
            saved to this connection.
          </Text>
          {preset === 'acr' && (
            <Banner theme="warn">
              ACR requires RBAC + ABAC repository permissions. Review existing
              role assignments with the registry owner before changing
              permission modes.
            </Banner>
          )}
          {missingRegion && (
            <Banner theme="warn">
              This identity has no saved region. Create a connection with an
              explicit region to generate scoped repository commands.
            </Banner>
          )}
          <form
            className="flex max-w-2xl flex-col gap-6"
            onSubmit={(event) => {
              event.preventDefault()
              form.handleSubmit()
            }}
          >
            <FormErrorBanner
              error={error}
              fallback="Command generation failed"
            />
            <form.Field name="repositories">
              {(field) => (
                <RepositoryPolicyFields
                  field={field}
                  id={`repository-access-${connection.id}`}
                  preset={preset}
                  disabled={isPending || missingRegion}
                />
              )}
            </form.Field>
            <div className="flex justify-end">
              <Button
                type="submit"
                variant="primary"
                disabled={!canSubmit || isPending || missingRegion}
                tooltipProps={
                  !canSubmit || missingRegion
                    ? {
                        tipContent: missingRegion
                          ? 'Cannot generate commands — this connection has no saved region'
                          : 'Cannot generate commands — provide valid repository names',
                      }
                    : undefined
                }
              >
                {isPending ? (
                  <>
                    <Icon variant="Loading" /> Generating commands
                  </>
                ) : (
                  'Generate commands'
                )}
              </Button>
            </div>
          </form>
          {currentResult && (
            <div className="flex min-w-0 flex-col gap-6" aria-live="polite">
              <section
                className="flex min-w-0 flex-col gap-4 border-t pt-6"
                aria-label="Generated repository access"
              >
                <SectionHeader
                  title="Generated repository access"
                  description={`Review the policy and run the commands in ${preset === 'gar' ? 'Google Cloud' : preset === 'acr' ? 'Azure' : 'AWS'} to grant read access. Existing repository grants are preserved.`}
                />
                <Tabs
                  naturalHeight
                  tabs={{
                    policy: (
                      <CodeBlock language="json" showCopy wrapLongLines>
                        {JSON.stringify(
                          currentResult.permissionsPolicy,
                          null,
                          2
                        )}
                      </CodeBlock>
                    ),
                    commands: (
                      <CodeBlock language="bash" showCopy wrapLongLines>
                        {currentResult.permissionsCli}
                      </CodeBlock>
                    ),
                  }}
                />
              </section>
              <Expand
                id={`repository-component-${connection.id}`}
                heading="Use this repository in a component"
              >
                <Text variant="subtext" theme="neutral">
                  After granting access and verifying the connection, use an
                  existing image tag in your app config.
                </Text>
                <CodeBlock
                  language="toml"
                  showCopy
                >{`name = "acme-image"\ntype = "container_image"\n\n${example}\ntag = "latest"`}</CodeBlock>
              </Expand>
            </div>
          )}
          <Text variant="subtext" theme="neutral">
            Generating commands does not change cloud permissions. Image imports
            check whether the image can be read.
          </Text>
        </>
      )}
    </section>
  )
}
