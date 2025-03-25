import { NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { ADMIN_API_URL } from '@/utils'

export const GET = withApiAuthRequired(async () => {
  return fetch(`${ADMIN_API_URL}/v1/orgs/admin-features`).then((res) =>
    res.json().then((features) => NextResponse.json(features))
  )
})
