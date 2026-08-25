# Village Government Website - Vue.js Version

A modern, responsive village government website built with Vue 3, Vite, and TailwindCSS. This is a complete rewrite of the React implementation in Vue.js.

## Features

- **Responsive Design**: Fully responsive on mobile, tablet, and desktop devices
- **Multiple Pages**: 8 main routes covering all village information
- **Data Visualization**: Charts and statistics for village demographics and economy
- **Interactive Map**: Leaflet-based map showing important village locations
- **News Management**: Full news listing and detail pages
- **UMKM Showcase**: Display of local businesses and enterprises
- **Public Information**: PPID (Layanan Informasi Publik) document management
- **Modern UI**: Built with TailwindCSS utility classes

## Technology Stack

- **Frontend Framework**: Vue 3 (Composition API)
- **Build Tool**: Vite
- **Styling**: TailwindCSS
- **Routing**: Vue Router 4
- **Mapping**: Leaflet
- **Language**: JavaScript (ES6+)

## Project Structure

```
src/
├── components/
│   ├── layout/
│   │   ├── Navbar.vue
│   │   ├── Footer.vue
│   │   └── Layout.vue
│   └── common/
│       ├── LoadingSpinner.vue
│       ├── Card.vue
│       └── EmptyState.vue
├── pages/
│   ├── Home.vue
│   ├── Profil.vue
│   ├── Infografik.vue
│   ├── Peta.vue
│   ├── Berita.vue
│   ├── BeritaDetail.vue
│   ├── UMKM.vue
│   └── PPID.vue
├── services/
│   └── desaService.js
├── router/
│   └── index.js
├── App.vue
├── main.js
└── style.css
```

## Routes

- `/` - Home page with hero slider, stats, and latest news
- `/profil` - Village profile with history, vision, mission, and officials
- `/infografik` - Data visualization with demographics, economy, and education stats
- `/peta` - Interactive map with important village locations
- `/berita` - News listing page
- `/berita/:slug` - Individual news article detail page
- `/umkm` - Local businesses showcase
- `/ppid` - Public information disclosure documents

## Getting Started

### Prerequisites

- Node.js 16+ and npm

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

The development server will open at `http://localhost:5173`

## Data Service Layer

All data is managed through `src/services/desaService.js` which provides:

- `getHomeData()` - Home page data with stats and latest news
- `getProfilData()` - Village profile information
- `getInfografikData()` - Chart and statistics data
- `getPetaData()` - Map markers and coordinates
- `getBeritaList()` - List of all news articles
- `getBeritaBySlug(slug)` - Single news article by slug
- `getUMKMData()` - Local businesses data
- `getPPIDData()` - Public documents

Each function simulates a 1-second network delay for realistic development experience.

## Component Guidelines

### Component Size
- All components are kept under 150 lines
- Complex logic is extracted into separate components
- Reusable components are centralized

### State Management
- Uses Vue 3 Composition API with `ref` and `computed`
- Local state for component-specific data
- No global state management needed

### Loading States
- All pages display `LoadingSpinner` during data fetch
- Empty states handled gracefully with `EmptyState` component
- Error handling with console logging

## Styling

- **Framework**: TailwindCSS utility classes only
- **Colors**: Emerald (primary), Blue (secondary), Gray (neutral)
- **Responsive**: Mobile-first approach with md/lg breakpoints
- **Animations**: Smooth transitions and fade-in effects

## Key Features

### Home Page
- Auto-rotating hero slider with 5-second interval
- Village statistics cards
- Village head welcome section
- Latest news preview
- Explore sections with quick links

### Navigation
- Sticky navbar with active route highlighting
- Mobile hamburger menu
- Persistent footer with contact info

### News System
- Grid layout for news listing
- Individual article pages with full content
- Date formatting in Indonesian locale
- Category badges

### Map Integration
- Leaflet-based interactive map
- Multiple location markers
- Popup information on marker click
- Responsive map container

### Data Visualization
- Progress bars for demographics
- Sector breakdown charts
- Education level statistics
- Infrastructure facility counts

## Performance

- Vite provides instant HMR (Hot Module Replacement)
- Optimized production build with code splitting
- TailwindCSS tree-shaking removes unused styles
- Lazy loading ready for future enhancements

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Mobile browsers (iOS Safari, Chrome Mobile)

## Development Workflow

1. Start dev server: `npm run dev`
2. Make changes to components
3. Changes auto-reload in browser
4. Build for production: `npm run build`
5. Preview build: `npm run preview`

## Code Quality

- Functional components only (no class components)
- Descriptive variable and function names
- Clear separation of concerns
- Reusable component patterns
- Consistent file naming (PascalCase for components)

## Future Enhancements

- Search functionality
- Multi-language support
- Admin panel for content management
- Real backend API integration
- User authentication
- Comments on news articles
- Contact form with email integration

## License

This project is part of the Desa Sukamaju government website initiative.

## Support

For issues or questions, please contact the development team.
