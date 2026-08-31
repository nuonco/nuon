import { Block } from './Block'
import { useShell } from './shell-context'

export const Utilities = () => {
  const { toggleText } = useShell()

  return (
    <div className="flex flex-none items-center gap-4">
      <Block
        className="h-[16px] cursor-pointer opacity-60"
        title="Toggle wireframe text"
        icon="TextAaIcon"
        onClick={toggleText}
      />
      <Block
        className="h-[16px] cursor-pointer"
        title="Search"
        icon="MagnifyingGlassIcon"
      />
      <Block
        className="h-[16px] cursor-pointer"
        title="Account"
        icon="UserIcon"
        text="user@example.com"
      />
    </div>
  )
}
