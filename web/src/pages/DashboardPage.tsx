import { useState } from 'react'
import type { User } from '../types/auth'
import { DashboardLayout } from '../components/layout'
import { WelcomeSection } from '../features/dashboard/WelcomeSection'
import { DashboardCards } from '../features/dashboard/DashboardCards'
import { BasePlaylistsPage } from './BasePlaylistsPage'

interface DashboardPageProps {
  user: User
  onLogout: () => void
}

type DashboardView = 'main' | 'base-playlists'

export function DashboardPage({ user, onLogout }: DashboardPageProps) {
  const [currentView, setCurrentView] = useState<DashboardView>('main')

  const handleNavigateToBasePlaylists = () => {
    setCurrentView('base-playlists')
  }

  const handleBackToDashboard = () => {
    setCurrentView('main')
  }

  return (
    <DashboardLayout user={user} onLogout={onLogout} onNavigateHome={handleBackToDashboard}>
      {currentView === 'main' && (
        <>
          <WelcomeSection user={user} />
          <DashboardCards onNavigateToBasePlaylists={handleNavigateToBasePlaylists} />
        </>
      )}
      {currentView === 'base-playlists' && (
        <BasePlaylistsPage onBack={handleBackToDashboard} />
      )}
    </DashboardLayout>
  )
}