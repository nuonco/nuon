import type { FC } from 'react'
import { redirect } from "next/navigation"
import type { IPageProps, TOrg } from '@/types'
import { nueQueryData } from "@/utils"

const Stratus: FC<IPageProps> = async () => {
  const { data, error} = await nueQueryData<Array<TOrg>>({
    path: `orgs`
  })

  if (data) {
    redirect(`/stratus/${data?.at(0).id}`)
  }
  
  return <>{error?.error}</>
}

export default Stratus
