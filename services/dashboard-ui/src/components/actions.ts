'use server'

import { revalidatePath, revalidateTag } from 'next/cache'
import { cookies } from 'next/headers'

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

export async function setDashboardSidebarCookie(isOpen: boolean) {
  const cookieStore = await cookies()
  cookieStore.set('is-sidebar-open', Boolean(isOpen).toString())
}
