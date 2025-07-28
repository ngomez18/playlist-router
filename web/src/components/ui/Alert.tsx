import type { ReactNode } from 'react'

interface AlertProps {
  children: ReactNode
  type?: 'info' | 'success' | 'warning' | 'error'
  className?: string
}

export function Alert({ children, type = 'info' }: AlertProps) {
  const baseClasses = 'alert'
  const typeClasses = {
    info: 'alert-info',
    success: 'alert-success', 
    warning: 'alert-warning',
    error: 'alert-error'
  }

  const classes = `${baseClasses} ${typeClasses[type]} alert-soft`

  return (
    <div className={classes}>
      <span>{children}</span>
    </div>
  )
}