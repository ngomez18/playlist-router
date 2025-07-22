# PlaylistSync - Product Requirement Document

## Overview

**Product Name:** PlaylistSync  
**Version:** 1.0  
**Date:** July 2025  
**Team:** Solo developer with AI assistance

## Problem Statement

Music lovers struggle to maintain multiple themed playlists (genres, eras, moods) when discovering new music. Currently, users must manually add each new song to multiple relevant playlists, leading to:
- Inconsistent playlist maintenance
- Songs being forgotten in certain playlists
- Time-consuming manual organization
- Abandoned themed playlists due to maintenance overhead

## Product Vision

Create a streamlined tool that automatically distributes songs from user's "base" playlists to multiple themed child playlists based on pre-configured rules, making playlist management effortless and consistent.

## Target Users

**Primary:** Spotify Premium subscribers who maintain multiple playlists (3+ playlists, add 5+ songs weekly)
**Secondary:** Music enthusiasts with Premium subscriptions who want better organization but find manual management tedious

**Note:** Free Spotify users are unlikely to convert to paid playlist tools. Marketing and product development should focus exclusively on existing Premium subscribers.

## Core User Journey

1. User connects Spotify account to PlaylistSync
2. User creates or designates multiple "base" playlists in Spotify
3. For each base playlist, user creates child playlists with custom filtering rules
4. User adds songs to their base playlists (normal Spotify workflow)
5. PlaylistSync automatically detects new songs and distributes them to matching child playlists
6. All playlists stay synchronized in user's Spotify account

## Functional Requirements

### Authentication & Onboarding
- **MUST:** Spotify OAuth integration for account linking
- **MUST:** Alternative email/password registration with option to link Spotify later
- **MUST:** Guided onboarding flow explaining base playlist concept
- **SHOULD:** Option to import existing playlists as starting templates

### Playlist Management
- **MUST:** Support for multiple base playlists per user
- **MUST:** Each base playlist can have multiple child playlists
- **MUST:** Child playlist creation with comprehensive rule-based filtering
- **MUST:** Support for all Spotify audio features as filter criteria:
  - **Musical Qualities:** Energy, Danceability, Valence, Tempo
  - **Technical Attributes:** Acousticness, Instrumentalness, Loudness, Key & Mode, Time Signature
  - **Context:** Liveness, Speechiness, Duration, Popularity
  - **Basic Metadata:** Genre, Release year
- **MUST:** Artist and song exclusion filters (text input with validation)
- **MUST:** Pre-defined playlist templates (e.g., "2000s Rock", "Chill Vibes", "Workout Music")
- **MUST:** Playlist preview before creation
- **SHOULD:** Bulk actions (enable/disable multiple playlists)
- **COULD:** Child playlist templates for easy reuse across base playlists

### User Interface
- **MUST:** Dark theme, modern minimalistic design using Chakra UI
- **MUST:** Desktop-first, mobile-friendly responsive design
- **MUST:** Multi-step wizard for child playlist creation on mobile devices
- **MUST:** Collapsible sections for organizing audio feature filters
- **MUST:** Dashboard showing all base playlists and their children with sync status
- **SHOULD:** Visual representation of base playlist and child relationships
- **SHOULD:** Search functionality for playlists
- **COULD:** Advanced playlist analytics and visualizations (premium feature)

### Sync Engine
- **MUST:** Automatic detection of new songs in base playlists
- **MUST:** Real-time distribution to matching child playlists
- **MUST:** Sync status dashboard showing last sync time and song count
- **MUST:** Manual sync trigger option
- **SHOULD:** Sync history log (last 30 days)
- **SHOULD:** Error handling for failed syncs with retry mechanism
- **COULD:** Conflict resolution interface for songs that don't match any child playlist

### Pricing & Limits
- **MUST:** Free tier: 10 song distributions per month
- **MUST:** Basic tier ($0.99/month): Unlimited distributions, 2 base playlists, 5 children each
- **MUST:** Premium tier ($4.99/month): Unlimited distributions and playlists
- **MUST:** Usage tracking and limit enforcement
- **MUST:** Upgrade prompts when approaching limits
- **SHOULD:** Usage analytics for users to track their sync activity

**Pricing Research Note:** Most playlist/music organization tools are priced at $1-3/month. The $0.99 Basic tier is competitively positioned, while $4.99 Premium is at the higher end but justified by unlimited features.

## Technical Requirements

### Backend Architecture
- **Framework:** Go with Pocketbase for rapid development
- **Database:** SQLite (via Pocketbase) for user data, playlist configurations, and sync logs
- **Authentication:** JWT tokens with Pocketbase auth + Spotify OAuth
- **API Design:** Proper REST endpoints (private initially, mobile-ready)
- **API Integration:** Spotify Web API for playlist operations and audio features
- **Deployment:** Fly.io with unified backend + frontend deployment
- **Architecture:** Single Go binary serving React build + PocketBase APIs

### Frontend Architecture
- **Framework:** React 18 with TypeScript
- **Styling:** Chakra UI for component library and dark theme
- **State Management:** React Query for API state + Context for app state
- **Build Tool:** Vite for fast development and builds
- **Deployment:** Served from same Go binary as backend via Fly.io

### Code Organization
- **Repository:** Monorepo structure with Go backend and React frontend
- **Type Sharing:** Manual TypeScript/Go type definitions initially, with future automation via PocketBase TypeGen
- **Development:** Single deployment pipeline for integrated full-stack application

### External Integrations
- **Spotify Web API:** Primary integration for playlist management and audio features
- **Payment Processing:** Stripe for subscription management (future phase)

## Non-Functional Requirements

### Performance
- Page load times < 2 seconds
- API response times < 500ms for dashboard operations
- Sync operations complete within 30 seconds for 50 songs
- Mobile playlist creation wizard remains responsive

### Scalability
- Support 1000+ concurrent users initially
- Database queries optimized for multi-playlist operations
- Spotify API rate limit compliance (100 requests per minute)
- Efficient handling of multiple base playlists per user
- **Infrastructure costs:** $5-10/month initially, scaling to $400-700/month at 50K users (less than 1% of revenue at scale)

### Security
- HTTPS everywhere
- JWT token-based authentication
- Secure token storage for Spotify credentials
- Input validation and sanitization
- Rate limiting per user session
- Private API (no public access initially)

### Reliability
- 99.5% uptime target
- Graceful error handling with user-friendly messages
- Automatic retry logic for failed Spotify API calls
- Database backups (daily)

## Implementation Phases

### Phase 1: Core MVP (Weeks 1-4)
- Authentication system (email/password + Spotify OAuth)
- Single base playlist with basic child playlist creation
- Essential audio feature filters (energy, danceability, valence, genre, year)
- Manual sync functionality
- Basic dashboard with Chakra UI
- Free tier limits and usage tracking

### Phase 2: Multiple Base Playlists (Weeks 5-6)
- Support for multiple base playlists per user
- Enhanced dashboard showing base/child relationships
- All audio feature filters with collapsible UI organization
- Mobile-responsive multi-step wizard
- Sync history and detailed status tracking

### Phase 3: Advanced Features (Weeks 7-8)
- Artist and song exclusion filters
- Pre-defined playlist templates
- Automatic sync via polling (every 15 minutes)
- Error handling and retry logic
- Usage analytics dashboard

### Phase 4: Monetization (Weeks 9-10)
- Stripe integration
- Basic and Premium subscription tiers
- Payment management dashboard
- Tier-based limit enforcement

### Phase 5: Polish & Optimization (Weeks 11-12)
- Performance optimizations
- Advanced error handling
- User feedback integration
- Mobile experience refinements

## Success Metrics

### User Engagement
- Monthly Active Users (MAU)
- Average base playlists per user
- Average child playlists per base playlist
- Sync frequency per user (syncs per user per month)
- User retention (30-day, 90-day)
- **Conversion rate from Spotify Premium users** (key metric given target market)

### Business Metrics
- Conversion rate (free to basic to premium)
- Monthly Recurring Revenue (MRR)
- Customer Acquisition Cost (CAC)
- Churn rate by tier

### Technical Metrics
- API response times
- Sync success rate
- Error rates
- Spotify API rate limit adherence
- Mobile vs desktop usage patterns

## Risk Assessment

### Technical Risks
- **Spotify API changes:** Mitigate with proper error handling and API versioning
- **Rate limiting with multiple base playlists:** Implement intelligent queuing and user education
- **Mobile performance with complex filters:** Optimize UI and consider progressive loading
- **Scale issues:** Monitor performance and optimize database queries early

### Business Risks
- **Limited target market:** Only Premium Spotify users will realistically convert, significantly reducing TAM
- **Low conversion to paid tiers:** A/B test pricing and feature limits
- **High infrastructure costs:** Monitor usage patterns and optimize aggressively
- **Competition from Spotify:** Risk of Spotify building this functionality natively
- **Free browser extensions:** Competing with free alternatives that might satisfy basic needs

### Mitigation Strategies
- Focus on high-value features that justify subscription cost to existing Premium users
- Build comprehensive analytics from day one
- Implement feature flags for easy rollbacks
- Regular user feedback collection
- Careful monitoring of Spotify API usage across multiple playlists
- Differentiate from free alternatives through automation and advanced filtering

## Future Considerations

### Potential Features
- Support for Apple Music (limited API capabilities)
- Child playlist templates for reuse across base playlists
- Collaborative base playlists with shared rules
- AI-powered smart playlist suggestions
- Advanced analytics and playlist insights
- Social features (sharing playlist configurations)

### Technical Improvements
- Real-time sync via webhooks (if Spotify adds support)
- Local metadata caching for faster filtering
- Native mobile app using existing API
- Playlist backup and export functionality
- Advanced visualization of playlist relationships

## Conclusion

PlaylistSync addresses playlist management complexity by introducing a multi-base playlist workflow with comprehensive filtering capabilities. The technical architecture balances rapid development with future scalability, while the tiered pricing model ensures sustainable growth among Spotify Premium subscribers. 

Key strategic focus areas include targeting the Premium user segment exclusively, maintaining cost-effective infrastructure through unified deployment, and differentiating from free alternatives through advanced automation features. The phased implementation allows for iterative improvement based on user feedback while building toward a robust, profitable platform serving the most engaged segment of Spotify's user base.