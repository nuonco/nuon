import { Auth0Client } from '@auth0/nextjs-auth0/server'

export let auth0
if (typeof window === 'undefined') {
  auth0 = new Auth0Client({
    appBaseUrl: process.env.AUTH0_BASE_URL,
    authorizationParameters: {
      scope: 'openid profile email',
      audience: process.env.AUTH0_AUDIENCE,
    },
    clientId: process.env.AUTH0_CLIENT_ID,
    clientSecret: process.env.AUTH0_CLIENT_SECRET,
    domain: process.env.AUTH0_ISSUER_BASE_URL,
    routes: {
      login: '/api/auth/login',
      logout: '/api/auth/logout',
      callback: '/api/auth/callback',
      backChannelLogout: '/api/auth/backchannel-logout',
    },
    secret: process.env.AUTH0_SECRET,
    transactionCookie: {
      maxAge: 300,
    },
  })
}
