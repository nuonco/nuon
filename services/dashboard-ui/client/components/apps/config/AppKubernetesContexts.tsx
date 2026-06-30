import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TAppConfig } from '@/types'

export interface IAppKubernetesContexts {
  appConfig: TAppConfig
}

export const AppKubernetesContexts = ({ appConfig }: IAppKubernetesContexts) => {
  const contexts = appConfig?.kubernetes_contexts?.contexts ?? []

  if (contexts.length === 0) {
    return null
  }

  return (
    <div className="flex flex-col gap-3">
      {contexts.map((ctx) => {
        const sourceLink =
          ctx.org_id && ctx.app_id && ctx.source_component_id
            ? `/${ctx.org_id}/apps/${ctx.app_id}/components/${ctx.source_component_id}`
            : undefined

        return (
          <div
            key={ctx.id ?? ctx.name}
            className="flex flex-wrap gap-x-8 gap-y-2 items-start"
          >
            <LabeledValue label="Name">
              <Text family="mono" variant="subtext">
                {ctx.name}
              </Text>
            </LabeledValue>

            <LabeledValue label="Source component">
              <div className="flex items-center gap-1">
                <Icon variant="ArrowRightIcon" size={12} />
                {sourceLink ? (
                  <Link href={sourceLink} variant="default">
                    <Text family="mono" variant="subtext">
                      {ctx.source_component_name}
                    </Text>
                  </Link>
                ) : (
                  <Text family="mono" variant="subtext">
                    {ctx.source_component_name}
                  </Text>
                )}
              </div>
            </LabeledValue>
          </div>
        )
      })}
    </div>
  )
}
