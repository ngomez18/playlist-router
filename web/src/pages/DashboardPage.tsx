import type { User } from '../types/auth'
import { DashboardLayout } from '../components/layout'
import { WelcomeSection } from '../features/dashboard/WelcomeSection'
import { DashboardCards } from '../features/dashboard/DashboardCards'

interface DashboardPageProps {
  user: User
  onLogout: () => void
}

export function DashboardPage({ user, onLogout }: DashboardPageProps) {
  return (
    <DashboardLayout user={user} onLogout={onLogout}>
      <WelcomeSection user={user} />
      <DashboardCards />
    </DashboardLayout>
  )
}