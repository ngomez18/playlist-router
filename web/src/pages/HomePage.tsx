import { useEffect } from "react";
import { useAuth } from "../hooks/useAuth";
import { useAuthValidation } from "../hooks/useAuthValidation";
import { setAuthToken } from "../lib/auth";

function getInitials(name: string): string {
  return name
    .split(" ")
    .map((word) => word.charAt(0))
    .join("")
    .toUpperCase()
    .slice(0, 2); // Limit to 2 initials max
}

export default function HomePage() {
  const { isAuthenticated, user, logout } = useAuth();
  const { isLoading } = useAuthValidation();

  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || "";
  const loginUrl = `${apiBaseUrl}/auth/spotify/login`;

  // Handle OAuth callback - extract token from URL params
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get("token");

    if (token) {
      // Store token and clean URL
      setAuthToken(token);
      window.history.replaceState({}, document.title, window.location.pathname);
      // Reload to trigger auth validation
      window.location.reload();
    }
  }, []);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-base-100 flex items-center justify-center">
        <div className="loading loading-spinner loading-lg"></div>
      </div>
    );
  }

  if (isAuthenticated && user) {
    return (
      <div className="min-h-screen bg-base-100">
        <div className="navbar bg-base-200">
          <div className="flex-1">
            <a className="btn btn-ghost text-xl">PlaylistSync</a>
          </div>
          <div className="flex-none gap-2">
            <div className="dropdown dropdown-end">
              <div
                tabIndex={0}
                role="button"
                className="btn btn-ghost btn-circle avatar"
              >
                <div className="w-10 rounded-full bg-primary justify-center">
                  <span className="text-primary-content text-s font-bold">
                    {getInitials(user.name)}
                  </span>
                </div>
              </div>
              <ul
                tabIndex={0}
                className="mt-3 z-[1] p-2 shadow menu menu-sm dropdown-content bg-base-100 rounded-box w-52"
              >
                <li>
                  <span className="text-sm opacity-70">{user.email}</span>
                </li>
                <li>
                  <a onClick={logout}>Logout</a>
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div className="container mx-auto p-6">
          <div className="text-center mb-8">
            <h1 className="text-4xl font-bold mb-4">Welcome, {user.name}!</h1>
            <p className="text-lg">Ready to sync your playlists?</p>
          </div>

          {/* Dashboard content will go here */}
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            <div className="card bg-base-200 shadow-xl">
              <div className="card-body">
                <h2 className="card-title">Base Playlists</h2>
                <p>Manage your source playlists</p>
                <div className="card-actions justify-end">
                  <button className="btn btn-primary">View</button>
                </div>
              </div>
            </div>

            <div className="card bg-base-200 shadow-xl">
              <div className="card-body">
                <h2 className="card-title">Child Playlists</h2>
                <p>Configure filtered playlists</p>
                <div className="card-actions justify-end">
                  <button className="btn btn-primary">View</button>
                </div>
              </div>
            </div>

            <div className="card bg-base-200 shadow-xl">
              <div className="card-body">
                <h2 className="card-title">Sync History</h2>
                <p>View recent sync operations</p>
                <div className="card-actions justify-end">
                  <button className="btn btn-primary">View</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-base-100 flex items-center justify-center">
      <div className="text-center max-w-md">
        <h1 className="text-5xl font-bold mb-6">PlaylistSync</h1>
        <p className="text-lg mb-8">
          Automatically distribute songs from a base playlist to multiple themed
          child playlists based on configured rules.
        </p>
        <a href={loginUrl} className="btn btn-primary">
          <img src="/spotify-icon-dark.svg" alt="Spotify" className="w-5 h-5" />
          Log in with Spotify
        </a>
      </div>
    </div>
  );
}
