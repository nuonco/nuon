import { NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { nueQueryData } from '@/utils'

export const GET = withApiAuthRequired(async () => {
  const res = await nueQueryData({
    path: `orgs`,
  })

  return NextResponse.json(res)
})
