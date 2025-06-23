import { NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'

export const GET = async () => {
  const res = await nueQueryData({
    path: `orgs`,
  })

  return NextResponse.json(res)
}
