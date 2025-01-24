'use server'

import { revalidatePath, revalidateTag } from 'next/cache'

interface IRevalidateData {
  path?: string,
  tag?: string,
}

export async function revalidateData({ path, tag }: IRevalidateData) {
  if (path && !tag) {
      revalidatePath(path)
  }

  if (tag && !path) {
    revalidateTag(tag)
  }
}
