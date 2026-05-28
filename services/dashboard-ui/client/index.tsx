import React from 'react'
import ReactDOM from 'react-dom/client'
import { App } from './App'

const container = document.getElementById('root')!

const root = (globalThis.__reactRoot ??= ReactDOM.createRoot(container))
root.render(<App />)

if (import.meta.hot) {
  import.meta.hot.accept()
}
