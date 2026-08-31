import { ApprovalsButton } from './ApprovalsButton'
import { Block } from './Block'

export const StatusBar = () => (
  <footer className="mx-4 mb-4 flex flex-none items-center justify-between gap-4 rounded-lg bg-cool-grey-100/70 px-4 py-2 shadow-lg backdrop-blur-md dark:bg-dark-grey-800/70">
    <Block className="h-[12px] opacity-60" title="status" text="Connected" />
    <ApprovalsButton />
  </footer>
)
