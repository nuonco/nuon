import React from 'react'
import ReactDOM from 'react-dom/client'
import { App } from './App'
import { LiteApp } from './lite/LiteApp'
import { isLiteDashboard } from './lite/is-lite'

const container = document.getElementById('root')!

const root = (globalThis.__reactRoot ??= ReactDOM.createRoot(container))
root.render(isLiteDashboard() ? <LiteApp /> : <App />)

if (import.meta.hot) {
  import.meta.hot.accept()
}
