import { useState } from 'react'

export default function App() {
  const [status, setStatus] = useState('not checked')

  async function checkBackend() {
    try {
      const res = await fetch('http://localhost:8080/api/health')
      const data = await res.json()
      setStatus(data.status)
    } catch (e) {
      setStatus('error — is the backend running?')
    }
  }

  return (
    <div>
      <h1>MeetMRT</h1>
      <button onClick={checkBackend}>Check backend</button>
      <p>Status: {status}</p>
    </div>
  )
}