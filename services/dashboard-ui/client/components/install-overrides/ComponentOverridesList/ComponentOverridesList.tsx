import { EmptyState } from '@/components/common/EmptyState'
import { ComponentOverrideCard } from '@/components/install-overrides/ComponentOverrideCard'
import { groupComponentOverrideInputs } from '@/utils/install-utils'
import type { TAppInput, TComponentType } from '@/types'

export interface IComponentOverridesList {
  inputs?: TAppInput[]
  values?: Record<string, string>
  showEnabled?: boolean
  codeBlockVariant?: 'compact' | 'viewer'
  componentTypes?: Record<string, TComponentType>
  muteDisabled?: boolean
  typeVariant?: 'badge' | 'icon'
}

export const ComponentOverridesList = ({
  inputs,
  values,
  showEnabled = true,
  codeBlockVariant = 'compact',
  componentTypes,
  muteDisabled = false,
  typeVariant = 'badge',
}: IComponentOverridesList) => {
  const cards = groupComponentOverrideInputs(inputs || [])

  if (cards.length === 0) {
    return (
      <EmptyState
        emptyTitle="No component overrides"
        emptyMessage="This install has no per-component overrides."
        variant="diagram"
        size="sm"
      />
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {cards.map((card) => (
        <ComponentOverrideCard
          key={card.component}
          card={card}
          values={values}
          readOnly
          showEnabled={showEnabled}
          codeBlockVariant={codeBlockVariant}
          componentType={componentTypes?.[card.component]}
          muteDisabled={muteDisabled}
          typeVariant={typeVariant}
        />
      ))}
    </div>
  )
}
