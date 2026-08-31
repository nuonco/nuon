import { Block } from './Block'
import { Drawer } from './Drawer'
import { LinkBlock } from './LinkBlock'
import { labelWidth } from './utils'

export interface INodeDrawer {
  title: string
  path?: string
  onClose: () => void
}

const fields = ['Status', 'Type', 'Updated', 'Version']

export const NodeDrawer = ({ title, path, onClose }: INodeDrawer) => (
  <Drawer title={title} onClose={onClose}>
    <div className="flex flex-col gap-3">
      <Block className="h-[12px] w-[70%]" />
      <Block className="h-[8px] w-[45%] opacity-50" />
    </div>

    <div className="flex flex-col gap-4">
      {fields.map((label) => (
        <div key={label} className="flex items-center justify-between">
          <Block
            className="h-[8px] opacity-50"
            style={{ width: labelWidth(label) }}
            title={label}
            text={label}
          />
          <Block className="h-[10px] w-[96px]" />
        </div>
      ))}
    </div>

    {path && (
      <LinkBlock
        path={path}
        label={`View ${title.toLowerCase()}`}
        className="h-[32px] w-full"
        style={{ width: '100%' }}
      />
    )}
  </Drawer>
)
