export default function HomePage() {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || ''
  const loginUrl = `${apiBaseUrl}/auth/spotify/login`

  return (
    <div className="min-h-screen bg-base-100 flex items-center justify-center">
      <div className="text-center max-w-md">
        <h1 className="text-5xl font-bold mb-6">PlaylistSync</h1>
        <p className="text-lg mb-8">
          Automatically distribute songs from your "base" playlists to multiple themed "child" playlists based on configured rules.
        </p>
        <a href={loginUrl} className="btn btn-primary">
          <img src="/spotify-icon-dark.svg" alt="Spotify" className="w-5 h-5" />
          Log in with Spotify
        </a>
      </div>
    </div>
  )
}