import { MapPin, TrendingUp, TrendingDown, Users } from 'lucide-react';

interface VisitorLocation {
  city: string;
  country: string;
  code: string;
  visitors: number;
  change: number;
  flag: string;
}

// Mock visitor data with flag emojis
const visitorLocations: VisitorLocation[] = [
  { city: 'New York', country: 'United States', code: 'US', visitors: 3542, change: 12.5, flag: '🇺🇸' },
  { city: 'London', country: 'United Kingdom', code: 'GB', visitors: 2876, change: 8.3, flag: '🇬🇧' },
  { city: 'Tokyo', country: 'Japan', code: 'JP', visitors: 2345, change: -5.2, flag: '🇯🇵' },
  { city: 'Paris', country: 'France', code: 'FR', visitors: 1876, change: 15.7, flag: '🇫🇷' },
  { city: 'Berlin', country: 'Germany', code: 'DE', visitors: 1654, change: 3.1, flag: '🇩🇪' },
  { city: 'San Francisco', country: 'United States', code: 'US', visitors: 1432, change: 18.9, flag: '🇺🇸' },
  { city: 'Sydney', country: 'Australia', code: 'AU', visitors: 1234, change: 7.2, flag: '🇦🇺' },
  { city: 'Singapore', country: 'Singapore', code: 'SG', visitors: 987, change: 22.1, flag: '🇸🇬' },
  { city: 'Toronto', country: 'Canada', code: 'CA', visitors: 876, change: 11.4, flag: '🇨🇦' },
  { city: 'São Paulo', country: 'Brazil', code: 'BR', visitors: 654, change: 14.2, flag: '🇧🇷' },
];

export function VisitorLocations({ className = '' }: { className?: string }) {
  const totalVisitors = visitorLocations.reduce((sum, loc) => sum + loc.visitors, 0);
  const totalCountries = new Set(visitorLocations.map(loc => loc.code)).size;

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

      {/* World Map Background */}
      <div className="relative bg-gradient-to-b from-blue-50 to-blue-100 dark:from-gray-800 dark:to-gray-900 rounded-lg border border-[rgb(var(--border-primary))] p-4 h-48">
        <div className="absolute inset-0 opacity-10">
          <div className="flex items-center justify-center h-full">
            <div className="text-6xl">🌍</div>
          </div>
        </div>

        {/* Floating Stats */}
        <div className="relative h-full flex flex-col justify-between">
          <div className="flex justify-between items-start">
            <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur rounded-lg px-3 py-2 shadow-sm">
              <div className="flex items-center gap-2">
                <Users className="h-4 w-4 text-[rgb(var(--accent-mid))]" />
                <div>
                  <div className="text-lg font-bold text-[rgb(var(--text-primary))]">
                    {totalVisitors.toLocaleString()}
                  </div>
                  <div className="text-xs text-[rgb(var(--text-secondary))]">Total Visitors</div>
                </div>
              </div>
            </div>
            <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur rounded-lg px-3 py-2 shadow-sm">
              <div className="text-lg font-bold text-[rgb(var(--text-primary))]">
                {totalCountries}
              </div>
              <div className="text-xs text-[rgb(var(--text-secondary))]">Countries</div>
            </div>
          </div>

          <div className="flex justify-end">
            <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur rounded-lg px-3 py-2 shadow-sm">
              <div className="flex items-center gap-2">
                <MapPin className="h-4 w-4 text-[rgb(var(--accent-mid))]" />
                <div>
                  <div className="text-lg font-bold text-[rgb(var(--text-primary))]">
                    {visitorLocations.length}
                  </div>
                  <div className="text-xs text-[rgb(var(--text-secondary))]">Cities</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Location List */}
      <div className="space-y-2">
        <div className="text-xs text-[rgb(var(--text-secondary))] mb-2">Top 5 locations</div>
        {visitorLocations.slice(0, 5).map((location, index) => (
          <div key={index} className="flex items-center justify-between py-2.5 px-3 rounded-lg hover:bg-[rgb(var(--bg-hover))] transition-colors">
            <div className="flex items-center gap-3">
              <div className="w-7 h-7 rounded-full bg-[rgb(var(--accent-glow))] flex items-center justify-center text-sm font-bold text-[rgb(var(--accent))] flex-shrink-0">
                {index + 1}
              </div>
              <div className="flex items-center gap-2 flex-1 min-w-0">
                <span className="text-2xl" title={location.country}>{location.flag}</span>
                <div className="min-w-0">
                  <div className="text-sm font-medium text-[rgb(var(--text-primary))] truncate">{location.city}</div>
                  <div className="text-xs text-[rgb(var(--text-secondary))] truncate">{location.country}</div>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <div className="text-right">
                <div className="text-sm font-medium text-[rgb(var(--text-primary))]">{location.visitors.toLocaleString()}</div>
                <div className={`text-xs flex items-center justify-end gap-1 ${location.change >= 0 ? 'text-[rgb(var(--success))]' : 'text-[rgb(var(--error))]'}`}>
                  {location.change >= 0 ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                  {Math.abs(location.change)}%
                </div>
              </div>
              {/* Mini bar chart */}
              <div className="w-12 h-6 bg-[rgb(var(--bg-subtle))] rounded-full overflow-hidden">
                <div
                  className="h-full bg-[rgb(var(--accent-mid))] transition-all duration-500"
                  style={{ width: `${Math.min(100, (location.visitors / visitorLocations[0].visitors) * 100)}%` }}
                />
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="text-xs text-[rgb(var(--text-secondary))] border-t border-[rgb(var(--border-primary))] pt-2 flex items-center justify-between">
        <span>{visitorLocations.length} locations • {totalCountries} countries</span>
        <span>Updated 5 minutes ago</span>
      </div>
    </div>
  );
}