import type { ChangeEvent } from 'react'

interface InputProps {
  type?: 'text' | 'email' | 'password'
  placeholder?: string
  value?: string
  onChange?: (e: ChangeEvent<HTMLInputElement>) => void
  className?: string
  required?: boolean
  disabled?: boolean
}

export function Input({ 
  type = 'text',
  placeholder,
  value,
  onChange,
  className = '',
  required = false,
  disabled = false
}: InputProps) {
  const baseClasses = 'input input-bordered w-full'
  const classes = `${baseClasses} ${className}`

  return (
    <input
      type={type}
      placeholder={placeholder}
      value={value}
      onChange={onChange}
      className={classes}
      required={required}
      disabled={disabled}
    />
  )
}