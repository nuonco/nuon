'use client'

import { Card, type ICard } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TComponentConfig } from '@/types'

interface IComponentConfigCard extends Omit<ICard, 'children'> {
  config: TComponentConfig
}

export const ComponentConfigCard = ({
  config,
  ...props
}: IComponentConfigCard) => {
  const { install } = useInstall()
  const { org } = useOrg()

  return (
    <Card {...props}>
      <div className="flex flex-wrap items-start gap-4 justify-between">
        <HeadingGroup>
          <Text weight="strong">Component configuration</Text>
          <ID>{config.id}</ID>
        </HeadingGroup>

        <Text variant="subtext">
          <Link
            href={`/${org.id}/apps/${install.app_id}/configs/${install.app_config_id}/components/${config?.component_id}`}
          >
            View details <Icon variant="CaretRight" />
          </Link>
        </Text>
      </div>

      <div className="grid gap-6 md:grid-cols-4">
        <LabeledValue label="Version">{config?.version}</LabeledValue>
        <LabeledValue label="Type">{config?.type}</LabeledValue>
      </div>
    </Card>
  )
}

export const ComponentConfigCardSkeleton = (props: Omit<ICard, 'children'>) => {
  return (
    <Card {...props}>
      <div className="flex flex-wrap items-center gap-4">
        <Skeleton height="24px" width="106px" />
        <Skeleton height="17px" width="85px" />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <LabeledValue label={<Skeleton height="17px" width="34px" />}>
          <Skeleton height="23px" width="75px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="68px" />}>
          <Skeleton height="23px" width="110px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="41px" />}>
          <Skeleton height="23px" width="50px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="45px" />}>
          <Skeleton height="23px" width="54px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="53px" />}>
          <Skeleton height="23px" width="148px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="53px" />}>
          <Skeleton height="23px" width="215px" />
        </LabeledValue>
      </div>
    </Card>
  )
}
