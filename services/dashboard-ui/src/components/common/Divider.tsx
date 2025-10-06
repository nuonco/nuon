import { Text } from './Text'

export const Divider = ({ dividerWord }: { dividerWord?: string }) => {
  return (
    <div className="relative">
      <hr />
      {dividerWord ? (
        <Text className="text-[11px] leading-[18px] shadow-sm px-2 border w-fit rounded-lg bg-white text-cool-grey-950 dark:bg-dark-grey-800 dark:text-cool-grey-50 absolute inset-0 m-auto h-[20px]">
          {dividerWord.toLocaleUpperCase()}
        </Text>
      ) : null}
    </div>
  )
}
