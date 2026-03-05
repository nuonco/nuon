interface PylonChatSettings {
  app_id: string
  email: string
  name?: string
  avatar_url?: string
  email_hash?: string
  account_id?: string
  account_external_id?: string
}

interface PylonConfig {
  chat_settings: PylonChatSettings
}

interface PylonAPI {
  (...args: unknown[]): void
  q: unknown[][]
  e: (args: unknown[]) => void
}

declare global {
  interface Window {
    pylon: PylonConfig
    Pylon: PylonAPI
  }
}

export {}
