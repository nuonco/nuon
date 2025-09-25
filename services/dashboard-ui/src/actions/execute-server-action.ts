'use server'

import { revalidatePath } from 'next/cache'

export interface IServerAction {
  orgId: string
  path?: string
}

type TServerActionFn<TArgs, TResult> = (args: TArgs) => Promise<TResult>

export async function executeServerAction<TArgs, TResult>({
  action,
  args,
  path,
}: {
  action: TServerActionFn<TArgs, TResult>
  args: TArgs
  path?: string
}): Promise<TResult> {
  const result = await action(args)
  if (path) revalidatePath(path)
  return result
}
