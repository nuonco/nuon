export default {
  title: 'Playground / Installs / Activity triggered by filter',
}

import { useState } from 'react'
import { ActorFilterPreview, TriggeredByDropdown } from './ActorFilter.preview'

export const InteractiveFeed = () => (
  <div className="max-w-4xl mx-auto p-6">
    <ActorFilterPreview />
  </div>
)

export const FilteredToTeammate = () => (
  <div className="max-w-4xl mx-auto p-6">
    <ActorFilterPreview initialActor="acc-1" />
  </div>
)

export const PickerOpen = () => {
  const [selected, setSelected] = useState<string>()
  return (
    <div className="flex justify-end p-6 pb-[420px] max-w-xl">
      <TriggeredByDropdown isOpen selected={selected} onChange={setSelected} />
    </div>
  )
}
