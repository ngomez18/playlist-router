# PlaylistSync Frontend

A React TypeScript frontend for the PlaylistSync application, built with Vite, Tailwind CSS, and DaisyUI.

## Tech Stack

- **React 18** with TypeScript
- **Vite** for fast development and building
- **Tailwind CSS** with **DaisyUI** for styling
- **React Router** for client-side routing

## Development

### Prerequisites

- Node.js 18+ 
- npm or yarn

### Getting Started

1. Install dependencies:
   ```bash
   npm install
   # or from the root directory:
   make frontend-install
   ```

2. Start the development server:
   ```bash
   npm run dev
   # or from the root directory:
   make frontend-dev
   ```

3. The frontend will be available at http://localhost:5173

### Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

### Project Structure

```
src/
├── components/       # Reusable UI components
├── pages/           # Page components
├── services/        # API service layer
├── App.tsx          # Main app component with routing
├── main.tsx         # App entry point
└── index.css        # Global styles (Tailwind imports)
```

## Features

### Current Pages

- **Home** (`/`) - Landing page with app introduction
- **Login** (`/login`) - Spotify OAuth login page
- **Dashboard** (`/dashboard`) - Main app interface for managing playlists

### Components

- **Layout** - Main app layout with navigation
- **Navbar** - Top navigation bar
- **PlaylistCard** - Individual playlist display card
- **CreatePlaylistModal** - Modal for creating new playlists

## Integration with Backend

The frontend is configured to work with the Go backend running on `http://localhost:8090`:

- API calls are made to `http://localhost:8090/api/*`
- Spotify OAuth redirects to `http://localhost:8090/auth/spotify/login`

## Development Workflow

To run both frontend and backend together:

```bash
# From the root directory
make dev-full
```

This will start:
- Backend at http://localhost:8090
- Frontend at http://localhost:5173

## Styling

The app uses DaisyUI themes on top of Tailwind CSS. Available themes:
- `light` (default)
- `dark`
- `cupcake`

To change themes, update the `data-theme` attribute on the HTML element or use DaisyUI's theme switching utilities.