import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useInstallLink } from '@/hooks/use-install-path'
import type { TAppConfig } from '@/types'

interface IImageConsumers {
  appConfig?: TAppConfig
  componentId: string
  componentHref: (componentId: string) => string
}

export const ImageConsumers = ({
  appConfig,
  componentId,
  componentHref,
}: IImageConsumers) => {
  const connections = appConfig?.component_config_connections ?? []
  const image = connections.find(
    (connection) => connection.component_id === componentId
  )
  if (!image?.component_name) return null

  const consumers = connections.filter(
    (connection) =>
      connection.component_id !== componentId &&
      connection.references?.some((reference) => {
        const [type, name] = reference.split('|')
        return type === 'component' && name === image.component_name
      })
  )
  if (!consumers.length) return null

  return (
    <div
      className="flex flex-wrap items-center gap-x-3 gap-y-1"
      data-testid="image-consumers"
    >
      <Text variant="subtext" theme="neutral">
        Used by
      </Text>
      {consumers.map((consumer) =>
        consumer.component_id ? (
          <Link
            key={consumer.component_id}
            href={componentHref(consumer.component_id)}
            className="inline-flex items-center gap-1.5 font-mono"
          >
            {consumer.type ? (
              <ComponentType
                type={consumer.type}
                displayVariant="icon-only"
                colorVariant="color"
                iconSize="14"
              />
            ) : null}
            {consumer.component_name ?? consumer.component_id}
          </Link>
        ) : null
      )}
    </div>
  )
}

export const ImageConsumersContainer = ({
  componentId,
}: Pick<IImageConsumers, 'componentId'>) => {
  const { appConfig } = useInstallAppConfig()
  const { install } = useInstall()
  const installLink = useInstallLink()
  return (
    <ImageConsumers
      appConfig={appConfig}
      componentId={componentId}
      componentHref={(id) =>
        installLink({
          installId: install?.id,
          appId: install?.app_id,
          suffix: `/components/${id}`,
        })
      }
    />
  )
}
