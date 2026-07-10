import { ComponentOverrideCard } from '@/components/install-overrides/ComponentOverrideCard'
import { groupComponentOverrideInputs } from '@/utils/install-utils'
import type { TAppInputConfig, TInstall } from '@/types'

export const ComponentOverridesSection = ({
  group,
  install,
  draftValues,
}: {
  group: NonNullable<TAppInputConfig['input_groups']>[number]
  install?: TInstall
  draftValues?: Record<string, string> | null
}) => {
  const installInputs = install ? install?.install_inputs?.at(0)?.values : {}

  const hasDraftValues = draftValues && Object.keys(draftValues).length > 0
  const normalizedDraftValues: Record<string, string> = {}
  if (hasDraftValues) {
    Object.entries(draftValues!).forEach(([key, value]) => {
      if (key.startsWith('inputs:')) {
        normalizedDraftValues[key.replace('inputs:', '')] = value
      }
    })
  }
  const mergedValues = hasDraftValues
    ? { ...installInputs, ...normalizedDraftValues }
    : installInputs

  const cards = groupComponentOverrideInputs(group?.app_inputs || [])
  if (cards.length === 0) {
    return null
  }

  return (
    <fieldset className="flex flex-col gap-6 border-t pt-6">
      <legend className="flex flex-col gap-0 pr-6">
        <span className="text-lg font-semibold">Components</span>
        <span className="text-sm font-normal">
          Choose which components deploy on this install and customize each one.
        </span>
      </legend>

      <div className="flex flex-col gap-4">
        {cards.map((card) => (
          <ComponentOverrideCard
            key={card.component}
            card={card}
            values={mergedValues}
          />
        ))}
      </div>
    </fieldset>
  )
}
