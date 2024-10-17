import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'


export const GET = withApiAuthRequired(async (_: NextRequest) => {
  let status = {
    status: 'error',
    status_description: 'Org health check deprecated',
  }
  return NextResponse.json(status)
})
