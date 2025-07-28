import type { ReactNode } from 'react'
import type { User } from '../../types/auth'
import { Navbar } from './'

interface DashboardLayoutProps {
  user: User
  onLogout: () => void
  children: ReactNode
}

export function DashboardLayout({ user, onLogout, children }: DashboardLayoutProps) {
  return (
    <div className="min-h-screen bg-base-100">
      <Navbar user={user} onLogout={onLogout} />
      <div className="container mx-auto p-6">
        {children}
      </div>
    </div>
  )
}