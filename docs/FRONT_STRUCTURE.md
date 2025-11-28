# PlaylistRouter Frontend Structure - Current Implementation

## Overview
**Status:** ✅ Implemented and deployed  
**Framework:** React 18 + TypeScript + Vite  
**UI Library:** Chakra UI with dark theme  
**State Management:** React Query + Context API  

## Directory Organization (Current)

```
web/src/
├── components/           # Reusable UI components
│   ├── layout/          # Layout components
│   │   ├── Layout.tsx   # Main app layout with navigation
│   │   └── Navbar.tsx   # Navigation bar component
│   └── ui/              # Basic UI components
├── features/            # Feature-specific components & logic
│   ├── auth/            # Authentication components
│   │   └── LoginPage.tsx # Spotify OAuth login page
│   ├── dashboard/       # Dashboard components
│   │   └── DashboardPage.tsx # Main dashboard view
│   └── playlists/       # Playlist management components
│       ├── BasePlaylistCard.tsx      # Base playlist display
│       ├── ChildPlaylistCard.tsx     # Child playlist display
│       ├── CreateChildPlaylistForm.tsx # Child playlist creation
│       ├── EditChildPlaylistForm.tsx   # Child playlist editing
│       └── PlaylistFilters.tsx       # Audio feature filters
├── hooks/               # Custom React hooks
│   └── useAuth.ts       # Authentication hook
├── lib/                 # Utilities and API clients
│   ├── api.ts          # API client with auth
│   └── audio-features.ts # Audio feature utilities
├── types/              # TypeScript type definitions
│   ├── auth.ts         # Auth-related types
│   ├── playlist.ts     # Playlist and filter types
│   └── spotify.ts      # Spotify API types
├── providers/          # React context providers
│   └── AuthProvider.tsx # Auth context provider
├── App.tsx             # Root component with routing
└── main.tsx           # Application entry point
```

## Component Architecture

### Layout Components (`components/layout/`)
- **Layout.tsx** - Main app wrapper with sidebar navigation
- **Navbar.tsx** - Top navigation with user menu and logout
- **Responsive design** - Mobile-first with collapsible sidebar

### Feature Components (`features/`)
- **Organized by domain** - auth, dashboard, playlists
- **Self-contained** - Each feature manages its own state and logic
- **Reusable patterns** - Form handling, API integration

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