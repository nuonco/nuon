import type { ReactNode } from 'react'
import { Block } from './Block'
import { labelWidth } from './utils'

export interface IDrawer {
  title: string
  onClose: () => void
  children: ReactNode
}

export const Drawer = ({ title, onClose, children }: IDrawer) => (
  <>
    <div
      className="fixed inset-0 z-40 bg-black/40"
      onClick={onClose}
      role="presentation"
    />
    <aside className="fixed inset-y-0 right-0 z-50 flex w-[420px] max-w-full flex-col gap-6 overflow-y-auto bg-background p-6">
      <header className="flex items-center justify-between gap-4">
        <Block
          className="h-[16px]"
          style={{ width: labelWidth(title) }}
          title={title}
          text={title}
        />
        <Block
          className="h-[16px] w-[16px] cursor-pointer"
          title="Close"
          icon="XIcon"
          onClick={onClose}
        />
      </header>
      {children}
    </aside>
  </>
)
