import { Block } from './Block'

export const LastUpdateTile = () => (
  <div className="flex flex-col gap-3 rounded-lg bg-cool-grey-100 p-4 dark:bg-dark-grey-800">
    <Block
      className="h-[10px] opacity-60"
      text="Last update"
      style={{ width: 88 }}
    />

    <div className="flex items-center gap-2">
      <Block
        className="h-[16px] flex-none opacity-80"
        icon="GitCommitIcon"
        iconSize={14}
        text="a1b2c3d"
      />
      <Block className="h-[10px] w-[110px] opacity-60" />
    </div>

    <div className="flex items-center gap-2">
      <Block className="h-[8px] w-[90px] opacity-40" />
      <Block className="h-[8px] w-[70px] opacity-40" />
    </div>
  </div>
)
