import { useState } from 'react'
import { InstallGraph, type IGraphNode } from './InstallGraph'

export default {
  title: 'Playground/Lite/InstallGraph',
}

export const Default = () => {
  const [selected, setSelected] = useState<IGraphNode | undefined>()

  return (
    <div className="p-4">
      <InstallGraph selectedId={selected?.id} onSelect={setSelected} />
    </div>
  )
}
