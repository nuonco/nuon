import { Block } from './Block'
import { labelWidth } from './utils'

export default {
  title: 'Playground/Lite/Block',
}

export const Default = () => <Block className="w-[200px] h-[32px]" />

export const Sizes = () => (
  <div className="flex flex-col gap-4">
    <Block className="w-[76px] h-[32px]" title="logo" />
    <Block className="w-[200px] h-[24px]" title="breadcrumb" />
    <Block className="w-full h-[4rem]" title="row" />
  </div>
)

export const DerivedFromLabel = () => (
  <div className="flex gap-4 items-center">
    {['Save', 'Create app', 'Add component'].map((label) => (
      <Block
        key={label}
        title={label}
        className="h-[32px]"
        style={{ width: labelWidth(label) }}
      />
    ))}
  </div>
)
