# Frontend Structure

## Directory Organization

```
src/
├── components/           # Reusable UI components
│   ├── ui/              # Basic UI components (Button, Card, Avatar, etc.)
│   ├── auth/            # Auth-specific components (LoginForm, UserMenu)
│   └── layout/          # Layout components (Navbar, DashboardLayout)
├── pages/               # Route-level page components
│   ├── AuthPage.tsx     # Login/landing page with OAuth callback handling
│   ├── DashboardPage.tsx # Main authenticated dashboard
│   └── HomePage.tsx     # Root page component (router logic)
├── features/            # Feature-specific components & logic
│   └── dashboard/       # Dashboard-specific components
├── hooks/               # Custom React hooks
├── lib/                 # Utilities and API clients
├── contexts/            # React contexts for global state
└── types/               # TypeScript type definitions
```

## Component Philosophy

### UI Components (`components/ui/`)
- **Reusable, generic components**
- **No business logic** - just UI presentation
- **Consistent API** with props for customization
- **DaisyUI integration** with Tailwind classes

### Auth Components (`components/auth/`)
- **Authentication-specific UI**
- **LoginForm** - Spotify login interface
- **UserMenu** - User avatar dropdown with logout

### Layout Components (`components/layout/`)
- **Structural page layout**
- **Navbar** - Top navigation with user menu
- **DashboardLayout** - Wrapper for authenticated pages

### Pages (`pages/`)
- **Route-level components**
- **Minimal logic** - mostly composition of smaller components
- **AuthPage** - Handles login and OAuth callback
- **DashboardPage** - Main authenticated interface
- **HomePage** - Root component with auth routing logic

### Features (`features/`)
- **Feature-specific components**
- **Business logic and domain-specific UI**
- **DashboardCards** - Playlist management cards
- **WelcomeSection** - Personalized welcome message

## Import Patterns

### Clean Imports with Index Files
```typescript
// ✅ Good - using index exports
import { Button, Card, Avatar } from '../components/ui'
import { LoginForm, UserMenu } from '../components/auth'

// ❌ Avoid - direct file imports
import { Button } from '../components/ui/Button'
import { Card } from '../components/ui/Card'
```

### Component Props Pattern
```typescript
// ✅ Consistent prop interfaces
interface ComponentProps {
  children?: ReactNode
  className?: string
  // ... component-specific props
}
```

## Benefits of This Structure

1. **Separation of Concerns** - UI, business logic, and routing are clearly separated
2. **Reusability** - UI components can be used across different features
3. **Maintainability** - Easy to find and modify specific functionality
4. **Scalability** - New features can be added without restructuring
5. **Testing** - Components can be tested in isolation
6. **Clean Imports** - Index files make imports cleaner and more organized

## Next Steps

When adding new features:

1. **Create feature directory** in `features/` (e.g., `features/playlists/`)
2. **Build reusable UI components** in `components/ui/`
3. **Create pages** for new routes in `pages/`
4. **Add business logic** in feature-specific components
5. **Update index files** for clean imports