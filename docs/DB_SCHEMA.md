# PlaylistRouter Database Schema - PocketBase Collections

## Overview

This document defines the complete database schema for PlaylistRouter using PocketBase collections. The design leverages PocketBase's built-in features while maintaining clean separation of concerns across authentication, integrations, and playlist management.

**Current Implementation Status:** Core collections implemented and deployed. Subscription/billing collections planned for future releases.

## Collection Architecture

**✅ Implemented Collections:**
```
┌─────────────────┐
│ users (built-in)│
└─────────┬───────┘
          │
          ├── spotify_integrations (1:1) ✅
          ├── base_playlists (1:many) ✅
          ├── child_playlists (1:many) ✅
          └── sync_events (1:many) ✅
                    │
                    └── base_playlists (many:1)
```

**🔮 Future Collections:**
- subscriptions (billing/usage tracking)
- usage_logs (detailed analytics)

---

## 1. Users Collection (Built-in)

**Collection Name:** `users`  
**Type:** Built-in PocketBase authentication collection  
**Purpose:** Core user authentication and profile management

### Schema
```typescript
interface User {
  id: string;              // Auto-generated UUID
  email: string;           // Unique, required
  emailVisibility: boolean; // PocketBase built-in
  verified: boolean;       // Email verification status
  username?: string;       // Optional, unique if provided
  created: Date;           // Auto-generated
  updated: Date;           // Auto-updated
}
```

### Access Rules
```javascript
listRule: ""      // Admin only
viewRule: "id = @request.auth.id"
createRule: ""    // Public registration handled by PocketBase
updateRule: "id = @request.auth.id"
deleteRule: "id = @request.auth.id"
```

### Indexes
- `email` (unique, built-in)
- `username` (unique, built-in)

---

## 2. Subscriptions Collection (FUTURE)

**Collection Name:** `subscriptions`  
**Purpose:** Manage user subscription tiers and billing integration with Stripe  
**Status:** 🔮 Planned for future release

### Schema
```typescript
interface Subscription {
  id: string;                    // Auto-generated UUID
  user: string;                  // Relation to users.id (required)
  
  // Stripe Integration
  stripe_subscription_id?: string; // Unique Stripe subscription ID
  stripe_customer_id?: string;     // Stripe customer ID
  
  // Subscription Details
  tier: 'basic' | 'premium';                              // Required
  status: 'active' | 'past_due' | 'canceled' | 'incomplete' | 'unpaid'; // Required
  current_period_start: Date;     // Billing period start (required)
  current_period_end: Date;       // Billing period end (required)
  cancel_at_period_end: boolean;  // Default: false
  
  // Timestamps
  created: Date;                  // Auto-generated
  updated: Date;                  // Auto-updated
}
```

### Field Validations
- `stripe_subscription_id`: Unique when provided
- `stripe_customer_id`: Required when `stripe_subscription_id` exists
- `current_period_end` must be after `current_period_start`
- Only one `active` subscription per user (enforced at application level)

### Access Rules
```javascript
listRule: "user = @request.auth.id"
viewRule: "user = @request.auth.id"
createRule: ""    // Backend only (via Stripe webhooks)
updateRule: ""    // Backend only (via Stripe webhooks)
deleteRule: ""    // Backend only
```

### Indexes
- `user` (for user subscription lookup)
- `stripe_subscription_id` (unique, for webhook processing)
- `status` (for billing operations)
- `current_period_end` (for expiration checks)

---

## 3. Sync Events Collection (IMPLEMENTED)

**Collection Name:** `sync_events`  
**Purpose:** Track sync operations and their results for auditing and debugging  
**Status:** ✅ Implemented and deployed

### Schema
```typescript
interface SyncEvent {
  id: string;                    // Auto-generated UUID
  user_id: string;               // Relation to users.id (required)
  base_playlist_id: string;      // Relation to base_playlists.id (required)
  child_playlist_ids?: string[]; // JSON array of affected child playlist IDs
  status: 'in_progress' | 'completed' | 'failed'; // Required
  started_at: Date;              // Required
  completed_at?: Date;           // Completion timestamp
  error_message?: string;        // Error details if failed
  tracks_processed: number;      // Number of tracks processed
  total_api_requests: number;    // API calls made during sync
  created: Date;                 // Auto-generated
  updated: Date;                 // Auto-updated
}
```

### Field Validations
- `status`: Must be one of the defined enum values
- `started_at`: Required timestamp
- `tracks_processed`: Default 0
- `total_api_requests`: Default 0

### Access Rules
```javascript
listRule: "user_id = @request.auth.id"
viewRule: "user_id = @request.auth.id"
createRule: ""    // Backend only
updateRule: ""    // Backend only
deleteRule: ""    // Backend only
```

### Indexes
- `user_id` (for user sync history lookup)
- `base_playlist_id` (for playlist sync history)
- `status` (for monitoring sync operations)
- `started_at` (for chronological ordering)

---

## 4. Spotify Integrations Collection (IMPLEMENTED)

**Collection Name:** `spotify_integrations`  
**Purpose:** Store Spotify OAuth tokens and user linking (one-to-one relationship with users)  
**Status:** ✅ Implemented and deployed

### Schema
```typescript
interface SpotifyIntegration {
  id: string;                  // Auto-generated UUID
  user: string;                // Relation to users.id (required, unique)
  
  // Spotify Account Details
  spotify_id: string;          // Spotify user ID (required, unique)
  display_name?: string;       // Spotify display name
  
  // OAuth Tokens (secured/encrypted by PocketBase)
  access_token: string;        // Required
  refresh_token: string;       // Required
  token_type: string;          // Default: "Bearer"
  expires_at: Date;            // Token expiration timestamp
  scope?: string;              // Granted permissions (space-separated)
  
  // Status
  is_active: boolean;          // Default: true
  
  // Timestamps
  created: Date;               // Auto-generated
  updated: Date;               // Auto-updated
}
```

### Field Validations
- `user`: Unique (one Spotify integration per user)
- `spotify_id`: Unique globally (one Spotify account per integration)
- `access_token` and `refresh_token`: Required for active integrations
- `expires_at`: Must be future date when creating/updating tokens

### Access Rules
```javascript
listRule: "user = @request.auth.id"
viewRule: "user = @request.auth.id"
createRule: "user = @request.auth.id"
updateRule: "user = @request.auth.id"
deleteRule: "user = @request.auth.id"
```

### Indexes
- `user` (unique, for user integration lookup)
- `spotify_id` (unique, for Spotify account lookup)
- `is_active` (for active integrations)
- `expires_at` (for token refresh operations)

---

## 5. Base Playlists Collection (IMPLEMENTED)

**Collection Name:** `base_playlists`  
**Purpose:** Track user's primary playlists that serve as sources for distribution  
**Status:** ✅ Implemented and deployed

### Schema
```typescript
interface BasePlaylist {
  id: string;                  // Auto-generated UUID
  user_id: string;             // Relation to users.id (required)
  
  // Playlist Details
  name: string;                // User-friendly name (required)
  spotify_playlist_id: string; // Spotify playlist ID (required)
  
  // Status
  is_active: boolean;          // Default: true
  
  // Timestamps
  created: Date;               // Auto-generated
  updated: Date;               // Auto-updated
}
```

**Note:** Sync status is now tracked in the separate `sync_events` collection for better auditing.
```

### Field Validations
- `spotify_playlist_id`: Unique per user (user can't add same playlist twice)
- `name`: 1-100 characters

### Access Rules
```javascript
listRule: "user_id = @request.auth.id"
viewRule: "user_id = @request.auth.id"
createRule: "user_id = @request.auth.id"
updateRule: "user_id = @request.auth.id"
deleteRule: "user_id = @request.auth.id"
```

### Indexes
- `user_id` (for user playlist lookup)
- `user_id + spotify_playlist_id` (unique combination)
- `is_active` (for active playlists)

---

## 6. Child Playlists Collection (IMPLEMENTED)

**Collection Name:** `child_playlists`  
**Purpose:** Store filtered playlists with rules for automatic song distribution  
**Status:** ✅ Implemented and deployed

### Schema
```typescript
interface ChildPlaylist {
  id: string;                  // Auto-generated UUID
  user_id: string;             // Relation to users.id (required)
  base_playlist_id: string;    // Relation to base_playlists.id (required)
  
  // Playlist Details
  name: string;                // User-friendly name (required)
  description?: string;        // Optional description
  spotify_playlist_id: string; // Spotify playlist ID (required)
  
  // Filtering Rules
  filter_rules?: MetadataFilters; // JSON object with metadata filtering
  
  // Status
  is_active: boolean;          // Default: true
  
  // Timestamps
  created: Date;               // Auto-generated
  updated: Date;               // Auto-updated
}

// Supporting Types (matches current implementation)
// Supporting Types (matches current implementation)
interface MetadataFilters {
  // Track Information
  duration_ms?: RangeFilter;
  popularity?: RangeFilter;
  explicit?: boolean;

  // Artist & Album Information
  genres?: SetFilter;
  release_year?: RangeFilter;
  artist_popularity?: RangeFilter;

  // Search-based Filters
  track_keywords?: SetFilter;
  artist_keywords?: SetFilter;
}

interface RangeFilter {
  min?: number;
  max?: number;
}

interface SetFilter {
  include?: string[];
  exclude?: string[];
}
```

### Field Validations
- `spotify_playlist_id`: Unique per user
- `name`: 1-100 characters
- `filter_rules`: Valid JSON conforming to MetadataFilters interface
- `base_playlist_id.user_id` must equal `user_id` (enforced via access rules)

### Access Rules
```javascript
listRule: "user_id = @request.auth.id"
viewRule: "user_id = @request.auth.id"
createRule: "user_id = @request.auth.id"
updateRule: "user_id = @request.auth.id"
deleteRule: "user_id = @request.auth.id"
```

### Indexes
- `user_id` (for user playlist lookup)
- `base_playlist_id` (for child playlist lookup)
- `user_id + spotify_playlist_id` (unique combination)
- `is_active` (for active playlists)

---

## Business Logic & Current Implementation

### Current Status
- **No subscription tiers implemented** - all users have unlimited access during MVP phase
- **No usage limits enforced** - focus on core functionality first
- **Sync tracking** - all operations logged in `sync_events` collection for future analytics

### Future Business Logic (Planned)
```typescript
// When subscriptions are implemented:
function getCurrentTier(userId: string): 'free' | 'basic' | 'premium' {
  // Will query subscriptions collection
  return 'free'; // Default for MVP
}

const FUTURE_TIER_LIMITS = {
  free: { monthly_syncs: 10, base_playlists: 0, child_playlists: 0 },
  basic: { monthly_syncs: Infinity, base_playlists: 2, child_playlists_per_base: 5 },
  premium: { monthly_syncs: Infinity, base_playlists: Infinity, child_playlists_per_base: Infinity }
};
```

---

## Data Relationships

### Implemented Relationships
#### One-to-Many Relationships
- `users` → `base_playlists` (user can have multiple base playlists)
- `users` → `child_playlists` (user can have multiple child playlists)
- `users` → `sync_events` (user can have multiple sync operations)
- `base_playlists` → `child_playlists` (base playlist can have multiple children)
- `base_playlists` → `sync_events` (base playlist can have multiple sync operations)

#### One-to-One Relationships
- `users` → `spotify_integrations` (user can have one Spotify integration)

### Current Constraints
- User can have only one Spotify integration (enforced by unique user relation)
- Each Spotify account can only be linked to one user (enforced by unique spotify_id)
- User cannot add the same Spotify playlist twice (as base or child) - enforced by unique indexes
- Child playlist must belong to same user as its base playlist (enforced by access rules)

### Future Constraints (Subscription Phase)
- User can have only one active subscription at a time
- Subscription tier limits on playlist counts and sync operations

---

## Migration & Scaling Considerations

### Future Schema Evolution
- **Playlist Templates**: Add `playlist_templates` collection for reusable filter configurations
- **Usage Tracking**: Add `monthly_usage` collection if detailed analytics needed
- **Sync Logs**: Add `sync_logs` collection for debugging and audit trails
- **Multi-provider**: Add additional integration collections for Apple Music, YouTube Music (e.g., `apple_music_integrations`)

### Performance Optimizations
- Consider materialized views for complex playlist relationship queries
- Cache frequently accessed data (user tier, playlist counts)
- Implement pagination for users with many playlists
- Background sync processing to avoid blocking user operations

### Security Considerations
- OAuth tokens are automatically secured by PocketBase field encryption
- Implement rate limiting per user for API operations
- Audit trail for subscription changes and billing events
- Secure webhook endpoints for Stripe integration
- Unique constraints prevent account linking conflicts