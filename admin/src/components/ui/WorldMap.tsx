import { useState } from 'react';
import { MapPin, TrendingUp, TrendingDown } from 'lucide-react';

interface VisitorData {
  country: string;
  code: string;
  visitors: number;
  change: number;
  path?: string; // SVG path for the country
}

// Visitor data by country
const visitorData: Record<string, VisitorData> = {
  US: { country: 'United States', code: 'US', visitors: 8934, change: 12.5 },
  GB: { country: 'United Kingdom', code: 'GB', visitors: 4521, change: 8.3 },
  JP: { country: 'Japan', code: 'JP', visitors: 3423, change: -5.2 },
  FR: { country: 'France', code: 'FR', visitors: 2987, change: 15.7 },
  DE: { country: 'Germany', code: 'DE', visitors: 2654, change: 3.1 },
  CA: { country: 'Canada', code: 'CA', visitors: 1876, change: 11.4 },
  AU: { country: 'Australia', code: 'AU', visitors: 1654, change: 7.2 },
  SG: { country: 'Singapore', code: 'SG', visitors: 1234, change: 22.1 },
  BR: { country: 'Brazil', code: 'BR', visitors: 987, change: 14.2 },
  IN: { country: 'India', code: 'IN', visitors: 2345, change: 28.3 },
  CN: { country: 'China', code: 'CN', visitors: 3456, change: -8.5 },
  RU: { country: 'Russia', code: 'RU', visitors: 876, change: -12.3 },
  MX: { country: 'Mexico', code: 'MX', visitors: 765, change: 18.9 },
  ES: { country: 'Spain', code: 'ES', visitors: 1543, change: 9.7 },
  IT: { country: 'Italy', code: 'IT', visitors: 1321, change: 6.4 },
};

// Simple SVG world map paths (greatly simplified)
const worldMapPaths = {
  // North America
  US: 'M 140 140 L 220 140 L 220 190 L 140 190 Z',
  CA: 'M 120 100 L 140 100 L 140 140 L 120 140 Z',
  MX: 'M 130 200 L 150 200 L 150 220 L 130 220 Z',

  // South America
  BR: 'M 180 280 L 220 280 L 220 340 L 180 340 Z',
  AR: 'M 170 350 L 190 350 L 190 380 L 170 380 Z',

  // Europe
  GB: 'M 410 120 L 420 120 L 420 130 L 410 130 Z',
  FR: 'M 410 140 L 430 140 L 430 160 L 410 160 Z',
  DE: 'M 430 130 L 445 130 L 445 145 L 430 145 Z',
  IT: 'M 425 160 L 435 160 L 435 180 L 425 180 Z',
  ES: 'M 400 160 L 415 160 L 415 175 L 400 175 Z',
  RU: 'M 480 100 L 580 100 L 580 150 L 480 150 Z',

  // Asia
  CN: 'M 600 140 L 680 140 L 680 180 L 600 180 Z',
  IN: 'M 580 200 L 620 200 L 620 240 L 580 240 Z',
  JP: 'M 720 150 L 740 150 L 740 170 L 720 170 Z',
  SG: 'M 650 260 L 660 260 L 660 270 L 650 270 Z',

  // Africa
  ZA: 'M 440 340 L 460 340 L 460 360 L 440 360 Z',
  EG: 'M 450 200 L 470 200 L 470 215 L 450 215 Z',

  // Oceania
  AU: 'M 680 320 L 720 320 L 720 360 L 680 360 Z',
  NZ: 'M 740 360 L 755 360 L 755 375 L 740 375 Z',
};

// Top cities with coordinates
const topCities = [
  { name: 'New York', country: 'US', lat: 40.7, lng: -74.0, visitors: 3542 },
  { name: 'London', country: 'GB', lat: 51.5, lng: -0.1, visitors: 2876 },
  { name: 'Tokyo', country: 'JP', lat: 35.7, lng: 139.7, visitors: 2345 },
  { name: 'Paris', country: 'FR', lat: 48.9, lng: 2.3, visitors: 1876 },
  { name: 'Berlin', country: 'DE', lat: 52.5, lng: 13.4, visitors: 1654 },
];

export function WorldMap({ className = '' }: { className?: string }) {
  const [hoveredCountry, setHoveredCountry] = useState<string | null>(null);
  const [selectedCity, setSelectedCity] = useState<typeof topCities[0] | null>(null);

  // Calculate color intensity based on visitor count
  const getCountryColor = (visitors: number) => {
    const maxVisitors = Math.max(...Object.values(visitorData).map(d => d.visitors));
    const intensity = visitors / maxVisitors;
    if (intensity > 0.7) return 'rgb(var(--accent-mid))';
    if (intensity > 0.4) return 'rgb(var(--accent-start))';
    if (intensity > 0.2) return 'rgb(var(--accent-glow))';
    return 'rgba(var(--text-tertiary), 0.2)';
  };

  // Convert lat/lng to map coordinates
  const latLngToMap = (lat: number, lng: number) => {
    const x = ((lng + 180) / 360) * 800;
    const y = ((90 - lat) / 180) * 400;
    return { x, y };
  };

  return (
    <div className={`space-y-3 ${className}`}>
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-display text-sm text-[rgb(var(--text-primary))] mb-1">
            Global Visitor Distribution
          </h3>
          <p className="text-xs text-[rgb(var(--text-secondary))]">
            Visitors by country for the last 30 days
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-[rgb(var(--accent-mid))] animate-pulse"></div>
          <span className="text-xs text-[rgb(var(--text-secondary))]">Live data</span>
        </div>
      </div>

      {/* World Map Container */}
      <div className="relative bg-gradient-to-b from-blue-50 to-blue-100 dark:from-gray-800 dark:to-gray-900 rounded-lg border border-[rgb(var(--border-primary))] overflow-hidden">
        <svg viewBox="0 0 800 400" className="w-full h-48">
          {/* Ocean background */}
          <rect width="800" height="400" fill="rgba(59, 130, 246, 0.05)" />

          {/* Grid lines */}
          <g stroke="rgba(156, 163, 175, 0.1)" strokeWidth="0.5">
            {Array.from({ length: 9 }, (_, i) => (
              <line key={`v${i}`} x1={i * 100} y1={0} x2={i * 100} y2={400} />
            ))}
            {Array.from({ length: 5 }, (_, i) => (
              <line key={`h${i}`} x1={0} y1={i * 100} x2={800} y2={i * 100} />
            ))}
          </g>

          {/* Countries */}
          {Object.entries(worldMapPaths).map(([code, path]) => {
            const data = visitorData[code];
            if (!data) return null;

            return (
              <path
                key={code}
                d={path}
                fill={getCountryColor(data.visitors)}
                stroke="rgba(156, 163, 175, 0.3)"
                strokeWidth="1"
                className="cursor-pointer transition-all duration-200"
                onMouseEnter={() => setHoveredCountry(code)}
                onMouseLeave={() => setHoveredCountry(null)}
                style={{
                  filter: hoveredCountry === code ? 'brightness(1.2)' : 'none'
                }}
              />
            );
          })}

          {/* City markers */}
          {topCities.map((city, index) => {
            const { x, y } = latLngToMap(city.lat, city.lng);
            const size = Math.max(4, Math.min(8, city.visitors / 500));

            return (
              <g key={index}>
                <circle
                  cx={x}
                  cy={y}
                  r={size}
                  fill="rgb(var(--accent-mid))"
                  stroke="white"
                  strokeWidth="2"
                  className="cursor-pointer"
                  style={{
                    filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.2))'
                  }}
                  onMouseEnter={() => setSelectedCity(city)}
                  onMouseLeave={() => setSelectedCity(null)}
                />
                {selectedCity?.name === city.name && (
                  <circle
                    cx={x}
                    cy={y}
                    r={size + 4}
                    fill="none"
                    stroke="rgb(var(--accent-mid))"
                    strokeWidth="2"
                    className="animate-ping"
                  />
                )}
              </g>
            );
          })}
        </svg>

        {/* Country Tooltip */}
        {hoveredCountry && visitorData[hoveredCountry] && (
          <div className="absolute top-4 left-4 p-3 rounded-lg shadow-lg bg-[rgb(var(--bg-elevated))] border border-[rgb(var(--border-primary))] min-w-[200px] z-10">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-lg">
                {visitorData[hoveredCountry].code === 'US' && '🇺🇸'}
                {visitorData[hoveredCountry].code === 'GB' && '🇬🇧'}
                {visitorData[hoveredCountry].code === 'JP' && '🇯🇵'}
                {visitorData[hoveredCountry].code === 'FR' && '🇫🇷'}
                {visitorData[hoveredCountry].code === 'DE' && '🇩🇪'}
                {visitorData[hoveredCountry].code === 'CA' && '🇨🇦'}
                {visitorData[hoveredCountry].code === 'AU' && '🇦🇺'}
                {visitorData[hoveredCountry].code === 'SG' && '🇸🇬'}
                {visitorData[hoveredCountry].code === 'BR' && '🇧🇷'}
                {visitorData[hoveredCountry].code === 'IN' && '🇮🇳'}
                {visitorData[hoveredCountry].code === 'CN' && '🇨🇳'}
                {visitorData[hoveredCountry].code === 'RU' && '🇷🇺'}
                {visitorData[hoveredCountry].code === 'MX' && '🇲🇽'}
                {visitorData[hoveredCountry].code === 'ES' && '🇪🇸'}
                {visitorData[hoveredCountry].code === 'IT' && '🇮🇹'}
              </span>
              <div className="font-semibold text-[rgb(var(--text-primary))]">
                {visitorData[hoveredCountry].country}
              </div>
            </div>
            <div className="space-y-1">
              <div className="flex justify-between items-center">
                <span className="text-xs text-[rgb(var(--text-secondary))]">Visitors:</span>
                <span className="text-sm font-medium text-[rgb(var(--text-primary))]">
                  {visitorData[hoveredCountry].visitors.toLocaleString()}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-xs text-[rgb(var(--text-secondary))]">Change:</span>
                <div className="flex items-center gap-1">
                  {visitorData[hoveredCountry].change >= 0 ? (
                    <TrendingUp className="h-3 w-3 text-[rgb(var(--success))]" />
                  ) : (
                    <TrendingDown className="h-3 w-3 text-[rgb(var(--error))]" />
                  )}
                  <span className={`text-sm font-medium ${
                    visitorData[hoveredCountry].change >= 0 ? 'text-[rgb(var(--success))]' : 'text-[rgb(var(--error))]'
                  }`}>
                    {Math.abs(visitorData[hoveredCountry].change)}%
                  </span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* City Tooltip */}
        {selectedCity && (
          <div className="absolute top-4 right-4 p-3 rounded-lg shadow-lg bg-[rgb(var(--bg-elevated))] border border-[rgb(var(--border-primary))] min-w-[180px] z-10">
            <div className="flex items-center gap-2 mb-2">
              <MapPin className="h-4 w-4 text-[rgb(var(--accent-mid))]" />
              <div className="font-semibold text-[rgb(var(--text-primary))]">{selectedCity.name}</div>
            </div>
            <div className="text-xs text-[rgb(var(--text-secondary))] mb-1">{selectedCity.country}</div>
            <div className="text-sm font-medium text-[rgb(var(--text-primary))]">
              {selectedCity.visitors.toLocaleString()} visitors
            </div>
          </div>
        )}
      </div>

      {/* Legend */}
      <div className="flex items-center justify-between text-xs">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded" style={{ backgroundColor: 'rgba(var(--text-tertiary), 0.2)' }}></div>
            <span className="text-[rgb(var(--text-secondary))]">Low</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded bg-[rgb(var(--accent-glow))]"></div>
            <span className="text-[rgb(var(--text-secondary))]">Medium</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded bg-[rgb(var(--accent-start))]"></div>
            <span className="text-[rgb(var(--text-secondary))]">High</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded bg-[rgb(var(--accent-mid))]"></div>
            <span className="text-[rgb(var(--text-secondary))]">Very High</span>
          </div>
        </div>
        <div className="text-[rgb(var(--text-secondary))]">
          {Object.keys(visitorData).length} countries • {topCities.length} top cities
        </div>
      </div>

      {/* Top Cities List */}
      <div className="space-y-2">
        <div className="text-xs text-[rgb(var(--text-secondary))] mb-2">Top 5 cities</div>
        {topCities.map((city, index) => (
          <div key={index} className="flex items-center justify-between py-2 px-3 rounded-lg hover:bg-[rgb(var(--bg-hover))] transition-colors">
            <div className="flex items-center gap-3">
              <div className="w-6 h-6 rounded-full bg-[rgb(var(--accent-glow))] flex items-center justify-center text-xs font-bold text-[rgb(var(--accent))]">
                {index + 1}
              </div>
              <div>
                <div className="text-sm font-medium text-[rgb(var(--text-primary))]">{city.name}</div>
                <div className="text-xs text-[rgb(var(--text-secondary))]">{city.country}</div>
              </div>
            </div>
            <div className="text-sm font-medium text-[rgb(var(--text-primary))]">
              {city.visitors.toLocaleString()}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}