interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

export function LoadingSpinner({ size = 'lg', className = '' }: LoadingSpinnerProps) {
  const sizeClasses = {
    sm: 'loading-sm',
    md: 'loading-md',
    lg: 'loading-lg'
  }

  return (
    <div className="min-h-screen bg-base-100 flex items-center justify-center">
      <div className={`loading loading-spinner ${sizeClasses[size]} ${className}`}></div>
    </div>
  )
}