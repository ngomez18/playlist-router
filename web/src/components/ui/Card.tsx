import type { ReactNode } from 'react'

interface CardProps {
  children: ReactNode
  className?: string
}

export function Card({ children, className = '' }: CardProps) {
  return (
    <div className={`card bg-base-200 shadow-xl ${className}`}>
      {children}
    </div>
  )
}

interface CardBodyProps {
  children: ReactNode
  className?: string
}

export function CardBody({ children, className = '' }: CardBodyProps) {
  return (
    <div className={`card-body ${className}`}>
      {children}
    </div>
  )
}

interface CardTitleProps {
  children: ReactNode
  className?: string
}

export function CardTitle({ children, className = '' }: CardTitleProps) {
  return (
    <h2 className={`card-title ${className}`}>
      {children}
    </h2>
  )
}

interface CardActionsProps {
  children: ReactNode
  className?: string
}

export function CardActions({ children, className = '' }: CardActionsProps) {
  return (
    <div className={`card-actions justify-end ${className}`}>
      {children}
    </div>
  )
}