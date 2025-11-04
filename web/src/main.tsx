import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { fullStory } from "./lib/fullstory";

fullStory.init(import.meta.env.VITE_FULLSTORY_ORG_ID || "");

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
