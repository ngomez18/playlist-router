import { useAuth } from "../hooks/useAuth";
import { useAuthValidation } from "../hooks/useAuthValidation";
import { LoadingSpinner } from "../components/ui";
import { AuthPage } from "./AuthPage";
import { DashboardPage } from "./DashboardPage";

export default function HomePage() {
  const { isAuthenticated, user, logout } = useAuth();
  const { isLoading } = useAuthValidation();

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (isAuthenticated && user) {
    return <DashboardPage user={user} onLogout={logout} />;
  }

  return <AuthPage />;
}
