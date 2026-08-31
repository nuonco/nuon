import { NavBlock } from './NavBlock'
import { primaryNav } from './nav'

export default {
  title: 'Playground/Lite/NavBlock',
}

export const Default = () => (
  <div className="flex flex-col gap-4 w-[240px]">
    {primaryNav.map((item) => (
      <NavBlock key={item.path} {...item} className="w-full h-[32px]" />
    ))}
  </div>
)

export const Horizontal = () => (
  <div className="flex gap-6 items-center">
    {primaryNav.map((item) => (
      <NavBlock key={item.path} {...item} exact className="w-[80px] h-[20px]" />
    ))}
  </div>
)
