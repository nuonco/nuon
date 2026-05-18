import { Button } from '@/components/common/Button'
import { Card, type ICard } from '@/components/common/Card'
import { Cron } from '@/components/common/Cron'
import { GitRepo } from '@/components/common/GitRepo'
import { Hash } from '@/components/common/Hash'
import { LabeledValue } from '@/components/common/LabeledValue'
import { OperationRolesList } from '@/components/common/OperationRolesList'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import {
  HelmValuesFilesModal,
  HelmValuesModal,
} from '@/components/components/configs/HelmConfig'
import { KubernetesManifestModal } from '@/components/components/configs/KubernetesConfig'
import {
  TerraformVariablesFilesModal,
  TerraformVariablesModal,
} from '@/components/components/configs/TerraformConfig'
import {
  PulumiConfigModal,
  PulumiEnvVarsModal,
} from '@/components/components/configs/PulumiConfig'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TComponentConfig } from '@/types'
import { getComponentConfigDisplayData } from '@/utils/component-config-display'

interface IComponentConfigCard extends Omit<ICard, 'children'> {
  config: TComponentConfig
  footer?: React.ReactNode
  headerActions?: React.ReactNode
}

export const ComponentConfigCard = ({
  config,
  footer,
  headerActions,
  ...props
}: IComponentConfigCard) => {
  const { commonFields, typeSpecificFields, vcsInfo, operationRoles } =
    getComponentConfigDisplayData(config)
  const { addModal } = useSurfaces()

  const getConfigButtons = () => {
    const buttons: Array<{ label: string; onClick: () => void }> = []

    switch (config.type) {
      case 'helm_chart':
        if (
          config.helm?.helm_config_json?.values &&
          Object.keys(config.helm.helm_config_json.values).length > 0
        ) {
          buttons.push({
            label: 'View values',
            onClick: () => {
              const modal = (
                <HelmValuesModal
                  values={config.helm!.helm_config_json!.values!}
                />
              )
              addModal(modal)
            },
          })
        }
        if (
          config.helm?.helm_config_json?.values_files &&
          config.helm.helm_config_json.values_files.length > 0
        ) {
          buttons.push({
            label: 'View values files',
            onClick: () => {
              const modal = (
                <HelmValuesFilesModal
                  valuesFiles={config.helm!.helm_config_json!.values_files!}
                />
              )
              addModal(modal)
            },
          })
        }
        break

      case 'terraform_module':
        if (
          config.terraform_module?.variables &&
          Object.keys(config.terraform_module.variables).length > 0
        ) {
          buttons.push({
            label: 'View variables',
            onClick: () => {
              const modal = (
                <TerraformVariablesModal
                  variables={config.terraform_module!.variables!}
                />
              )
              addModal(modal)
            },
          })
        }
        if (
          config.terraform_module?.variables_files &&
          config.terraform_module.variables_files.length > 0
        ) {
          buttons.push({
            label: 'View variables files',
            onClick: () => {
              const modal = (
                <TerraformVariablesFilesModal
                  variablesFiles={config.terraform_module!.variables_files!}
                />
              )
              addModal(modal)
            },
          })
        }
        break

      case 'kubernetes_manifest':
        if (config.kubernetes_manifest?.manifest) {
          buttons.push({
            label: 'View manifest',
            onClick: () => {
              const modal = (
                <KubernetesManifestModal
                  manifest={config.kubernetes_manifest!.manifest!}
                />
              )
              addModal(modal)
            },
          })
        }
        break

      case 'pulumi':
        if (
          config.pulumi?.config &&
          Object.keys(config.pulumi.config).length > 0
        ) {
          buttons.push({
            label: 'View config',
            onClick: () => {
              const modal = (
                <PulumiConfigModal
                  config={config.pulumi!.config!}
                />
              )
              addModal(modal)
            },
          })
        }
        if (
          config.pulumi?.env_vars &&
          Object.keys(config.pulumi.env_vars).length > 0
        ) {
          buttons.push({
            label: 'View env vars',
            onClick: () => {
              const modal = (
                <PulumiEnvVarsModal
                  envVars={config.pulumi!.env_vars!}
                />
              )
              addModal(modal)
            },
          })
        }
        break
    }

    return buttons
  }

  const configButtons = getConfigButtons()

  return (
    <Card {...props}>
      <div className="flex flex-col gap-6">
        <div className="flex items-center justify-between">
          <Text weight="strong">Configuration</Text>
          {(configButtons.length > 0 || headerActions) && (
            <div className="flex gap-2">
              {headerActions}
              {configButtons.map((button) => (
                <Button
                  key={button.label}
                  variant="secondary"
                  size="sm"
                  onClick={button.onClick}
                >
                  {button.label}
                </Button>
              ))}
            </div>
          )}
        </div>

        <div className="flex gap-6 items-start">
          <LabeledValue label="Version">{commonFields.version}</LabeledValue>
          <LabeledValue label="Type">
            <ComponentType type={config.type || 'unknown'} variant="subtext" />
          </LabeledValue>
        </div>

        <div className="grid gap-6 grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          {commonFields.buildTimeout && (
            <LabeledValue label="Build timeout">
              {commonFields.buildTimeout}
            </LabeledValue>
          )}
          {commonFields.deployTimeout && (
            <LabeledValue label="Deploy timeout">
              {commonFields.deployTimeout}
            </LabeledValue>
          )}
          {commonFields.driftSchedule && (
            <LabeledValue label="Drift schedule">
              <Cron cron={commonFields.driftSchedule} variant="subtext" />
            </LabeledValue>
          )}
          {commonFields.checksum && (
            <LabeledValue label="Checksum">
              <Hash hash={commonFields.checksum} />
            </LabeledValue>
          )}
          {typeSpecificFields.map((field) => (
            <LabeledValue key={field.label} label={field.label}>
              {field.value || '—'}
            </LabeledValue>
          ))}
        </div>

        {operationRoles &&
          Object.keys(operationRoles).length > 0 && (
            <div className="flex flex-col gap-2">
              <Text variant="body" weight="strong" level={5}>
                Operation Roles
              </Text>
              <OperationRolesList
                operationRoles={operationRoles}
              />
            </div>
          )}

        {vcsInfo?.vcsConfig && (
          <div className="pt-6 border-t">
            <div className="w-fit">
              <GitRepo vcsConfig={vcsInfo.vcsConfig} isAutoGrid />
            </div>
          </div>
        )}

        {footer && (
          <div className="pt-6 border-t flex flex-col gap-6">
            {footer}
          </div>
        )}
      </div>
    </Card>
  )
}

export const ComponentConfigCardSkeleton = (props: Omit<ICard, 'children'>) => {
  return (
    <Card {...props}>
      <div className="flex flex-col gap-6">
        <div className="flex flex-wrap items-center gap-4">
          <Skeleton height="24px" width="200px" />
          <Skeleton height="17px" width="120px" />
        </div>

        <div className="flex gap-6 items-start justify-start flex-wrap">
          <div className="space-y-2">
            <Skeleton height="20px" width="200px" />
            <Skeleton height="20px" width="150px" />
          </div>

          <LabeledValue label={<Skeleton height="17px" width="50px" />}>
            <Skeleton height="23px" width="30px" />
          </LabeledValue>

          <LabeledValue label={<Skeleton height="17px" width="40px" />}>
            <Skeleton height="23px" width="100px" />
          </LabeledValue>

          <LabeledValue label={<Skeleton height="17px" width="80px" />}>
            <Skeleton height="23px" width="50px" />
          </LabeledValue>

          <LabeledValue label={<Skeleton height="17px" width="90px" />}>
            <Skeleton height="23px" width="120px" />
          </LabeledValue>

          <LabeledValue label={<Skeleton height="17px" width="70px" />}>
            <Skeleton height="23px" width="80px" />
          </LabeledValue>
        </div>
      </div>
    </Card>
  )
}
