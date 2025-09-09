import { NextResponse } from 'next/server'
import { ADMIN_API_URL } from '@/configs/api'

export const GET = async () => {
  return fetch(`${ADMIN_API_URL}/v1/orgs/admin-features`).then((res) =>
    res.json().then((features) => NextResponse.json(features))
  )
}
