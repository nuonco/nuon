import { useState } from 'react'
import { Block } from './Block'
import { Drawer } from './Drawer'
import { PlaceholderGrid } from './PlaceholderGrid'

export default {
  title: 'Playground/Lite/Drawer',
}

export const Default = () => {
  const [isOpen, setIsOpen] = useState(true)

  return (
    <div className="p-4">
      <Block
        className="h-[32px] w-[120px] cursor-pointer"
        title="open drawer"
        onClick={() => setIsOpen(true)}
      />
      {isOpen && (
        <Drawer title="Component details" onClose={() => setIsOpen(false)}>
          <PlaceholderGrid rows={6} height="h-[2.5rem]" />
        </Drawer>
      )}
    </div>
  )
}
