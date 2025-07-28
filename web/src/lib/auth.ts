const TOKEN_KEY = 'auth_token'

export const getAuthToken = () => localStorage.getItem(TOKEN_KEY)

export const setAuthToken = (token: string) => localStorage.setItem(TOKEN_KEY, token)

export const removeAuthToken = () => localStorage.removeItem(TOKEN_KEY)