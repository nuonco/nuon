import { Block } from './Block'

export const BranchTile = () => (
  <div className="flex flex-col gap-3 rounded-lg bg-cool-grey-100 p-4 dark:bg-dark-grey-800">
    <Block
      className="h-[10px] opacity-60"
      text="Branch"
      style={{ width: 64 }}
    />

    <div className="flex items-center gap-2">
      <Block
        className="h-[16px] flex-none"
        icon="AppWindowIcon"
        iconSize={14}
        text="platform"
      />
      <Block
        className="h-[16px] flex-none rounded-full"
        icon="GitBranchIcon"
        iconSize={12}
        text="main"
      />
      <Block className="h-[16px] w-[56px] flex-none rounded-full" />
    </div>

    <Block
      className="h-[8px] flex-none opacity-50"
      icon="GitHub"
      iconSize={12}
      text="acme/platform"
    />
  </div>
)
