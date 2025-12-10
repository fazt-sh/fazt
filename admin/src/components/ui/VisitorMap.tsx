import { useState } from 'react';
import { MapPin, TrendingUp, TrendingDown } from 'lucide-react';

interface VisitorLocation {
  city: string;
  country: string;
  code: string;
  visitors: number;
  change: number;
  lat: number;
  lng: number;
}

// Mock visitor data
const visitorLocations: VisitorLocation[] = [
  { city: 'New York', country: 'United States', code: 'US', visitors: 3542, change: 12.5, lat: 40.7, lng: -74.0 },
  { city: 'London', country: 'United Kingdom', code: 'GB', visitors: 2876, change: 8.3, lat: 51.5, lng: -0.1 },
  { city: 'Tokyo', country: 'Japan', code: 'JP', visitors: 2345, change: -5.2, lat: 35.7, lng: 139.7 },
  { city: 'Paris', country: 'France', code: 'FR', visitors: 1876, change: 15.7, lat: 48.9, lng: 2.3 },
  { city: 'Berlin', country: 'Germany', code: 'DE', visitors: 1654, change: 3.1, lat: 52.5, lng: 13.4 },
  { city: 'San Francisco', country: 'United States', code: 'US', visitors: 1432, change: 18.9, lat: 37.8, lng: -122.4 },
  { city: 'Sydney', country: 'Australia', code: 'AU', visitors: 1234, change: 7.2, lat: -33.9, lng: 151.2 },
  { city: 'Singapore', country: 'Singapore', code: 'SG', visitors: 987, change: 22.1, lat: 1.3, lng: 103.8 },
  { city: 'Toronto', country: 'Canada', code: 'CA', visitors: 876, change: 11.4, lat: 43.7, lng: -79.4 },
  { city: 'São Paulo', country: 'Brazil', code: 'BR', visitors: 654, change: 14.2, lat: -23.5, lng: -46.6 },
];

export function VisitorMap({ className = '' }: { className?: string }) {
  const [selectedLocation, setSelectedLocation] = useState<VisitorLocation | null>(null);
  const totalVisitors = visitorLocations.reduce((sum, loc) => sum + loc.visitors, 0);

  return (
    <div className={`space-y-3 ${className}`}>
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-display text-sm text-[rgb(var(--text-primary))] mb-1">
            Top Visitor Locations
          </h3>
          <p className="text-xs text-[rgb(var(--text-secondary))]">
            Top 10 cities by visitor count
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-[rgb(var(--accent-mid))] animate-pulse"></div>
          <span className="text-xs text-[rgb(var(--text-secondary))]">Live data</span>
        </div>
      </div>

      {/* Simple World Map Visualization */}
      <div className="relative bg-[rgb(var(--bg-subtle))] rounded-lg border border-[rgb(var(--border-primary))] h-48 overflow-hidden">
        <div className="relative w-full h-full">
          {/* Map background with simplified world outline */}
          <svg viewBox="0 0 800 400" className="w-full h-full" preserveAspectRatio="xMidYMid meet">
            {/* Simple continent shapes */}
            <g fill="rgba(var(--text-tertiary), 0.05)" stroke="rgba(var(--text-tertiary), 0.2)" strokeWidth="1">
              {/* North America */}
              <ellipse cx="150" cy="140" rx="120" ry="80" />
              {/* South America */}
              <ellipse cx="220" cy="280" rx="60" ry="90" />
              {/* Europe */}
              <ellipse cx="400" cy="120" rx="60" ry="45" />
              {/* Africa */}
              <ellipse cx="400" cy="240" rx="75" ry="90" />
              {/* Asia */}
              <ellipse cx="560" cy="140" rx="160" ry="90" />
              {/* Australia */}
              <ellipse cx="640" cy="300" rx="60" ry="45" />
            </g>

            {/* Grid lines for reference */}
            <g stroke="rgba(var(--text-tertiary), 0.1)" strokeWidth="0.5">
              <line x1="0" y1="100" x2="800" y2="100" />
              <line x1="0" y1="200" x2="800" y2="200" />
              <line x1="0" y1="300" x2="800" y2="300" />
              <line x1="200" y1="0" x2="200" y2="400" />
              <line x1="400" y1="0" x2="400" y2="400" />
              <line x1="600" y1="0" x2="600" y2="400" />
            </g>

            {/* Location dots */}
            {visitorLocations.map((location, index) => {
              // Map lat/lng to SVG coordinates
              // lng: -180 to 180 maps to 0 to 800
              // lat: 90 to -90 maps to 0 to 400
              const x = ((location.lng + 180) / 360) * 800;
              const y = ((90 - location.lat) / 180) * 400;
              const size = Math.max(4, Math.min(12, location.visitors / 300));

              return (
                <g key={index}>
                  <circle
                    cx={x}
                    cy={y}
                    r={size}
                    fill="rgb(var(--accent-mid))"
                    fillOpacity="0.7"
                    stroke="rgb(var(--accent-mid))"
                    strokeWidth="1.5"
                    className="cursor-pointer transition-all duration-200"
                    style={{ filter: 'drop-shadow(0 0 4px rgba(var(--accent-mid), 0.5))' }}
                    onMouseEnter={() => setSelectedLocation(location)}
                    onMouseLeave={() => setSelectedLocation(null)}
                  />
                  {selectedLocation?.city === location.city && (
                    <>
                      <circle
                        cx={x}
                        cy={y}
                        r={size + 6}
                        fill="none"
                        stroke="rgb(var(--accent-mid))"
                        strokeWidth="2"
                        className="animate-ping"
                      />
                    </>
                  )}
                </g>
              );
            })}
          </svg>

          {/* Tooltip */}
          {selectedLocation && (
            <div className="absolute top-2 right-2 p-3 rounded-lg shadow-lg bg-[rgb(var(--bg-elevated))] border border-[rgb(var(--border-primary))] min-w-[180px] z-10">
              <div className="flex items-center gap-2 mb-2">
                <MapPin className="h-4 w-4 text-[rgb(var(--accent-mid))]" />
                <div>
                  <div className="font-semibold text-[rgb(var(--text-primary))]">{selectedLocation.city}</div>
                  <div className="text-xs text-[rgb(var(--text-secondary))]">{selectedLocation.country}</div>
                </div>
              </div>
              <div className="space-y-1">
                <div className="flex justify-between items-center">
                  <span className="text-xs text-[rgb(var(--text-secondary))]">Visitors:</span>
                  <span className="text-sm font-medium text-[rgb(var(--text-primary))]">{selectedLocation.visitors.toLocaleString()}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-xs text-[rgb(var(--text-secondary))]">Change:</span>
                  <div className="flex items-center gap-1">
                    {selectedLocation.change >= 0 ? (
                      <TrendingUp className="h-3 w-3 text-[rgb(var(--success))]" />
                    ) : (
                      <TrendingDown className="h-3 w-3 text-[rgb(var(--error))]" />
                    )}
                    <span className={`text-sm font-medium ${selectedLocation.change >= 0 ? 'text-[rgb(var(--success))]' : 'text-[rgb(var(--error))]'}`}>
                      {Math.abs(selectedLocation.change)}%
                    </span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Location List */}
      <div className="space-y-2">
        <div className="text-xs text-[rgb(var(--text-secondary))] mb-2">Top 5 locations</div>
        {visitorLocations.slice(0, 5).map((location, index) => (
          <div key={index} className="flex items-center justify-between py-2 px-3 rounded-lg hover:bg-[rgb(var(--bg-hover))] transition-colors cursor-pointer">
            <div className="flex items-center gap-3">
              <div className="w-6 h-6 rounded-full bg-[rgb(var(--accent-glow))] flex items-center justify-center text-xs font-bold text-[rgb(var(--accent))]">
                {index + 1}
              </div>
              <div>
                <div className="text-sm font-medium text-[rgb(var(--text-primary))]">{location.city}</div>
                <div className="text-xs text-[rgb(var(--text-secondary))]">{location.country}</div>
              </div>
            </div>
            <div className="text-right">
              <div className="text-sm font-medium text-[rgb(var(--text-primary))]">{location.visitors.toLocaleString()}</div>
              <div className={`text-xs flex items-center justify-end gap-1 ${location.change >= 0 ? 'text-[rgb(var(--success))]' : 'text-[rgb(var(--error))]'}`}>
                {location.change >= 0 ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                {Math.abs(location.change)}%
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="text-xs text-[rgb(var(--text-secondary))] border-t border-[rgb(var(--border-primary))] pt-2">
        {visitorLocations.length} locations • {totalVisitors.toLocaleString()} total visitors
      </div>
    </div>
  );
}