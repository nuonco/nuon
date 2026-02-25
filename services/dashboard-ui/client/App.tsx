import { BrowserRouter, Routes, Route } from 'react-router'

const Root = () => <>root</>
const Apps = () => <>apps</>

export const App = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Root />} />
        <Route path="/apps" element={<Apps />} />
      </Routes>
    </BrowserRouter>
  )
}
