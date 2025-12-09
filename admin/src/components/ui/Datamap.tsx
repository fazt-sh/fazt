import { useEffect, useRef, useState } from 'react';
import { visitorData, flagMap, mapStats } from '../../data/mapData';

interface CountryData {
  active: {
    value: string;
    percent: string;
    isGrown: boolean;
  };
  new: {
    value: string;
    percent: string;
    isGrown: boolean;
  };
  fillKey: 'LOW' | 'MEDIUM' | 'HIGH' | 'MAJOR';
  short: string;
  customName?: string;
}

declare global {
  interface Window {
    Datamap: {
      new(options: any): {
        resize(): void;
        updateChoropleth(data: any, options?: any): void;
      };
    };
    d3: any;
    topojson: any;
  }
}

export function Datamap() {
  const mapRef = useRef<HTMLDivElement>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);
  const mapInstanceRef = useRef<ReturnType<Window['Datamap']['new']> | null>(null);
  const resizeTimeoutRef = useRef<NodeJS.Timeout>();

  // Load scripts
  useEffect(() => {
    if (window.Datamap && window.d3 && window.topojson) {
      setIsLoaded(true);
      return;
    }

    const loadScript = (src: string): Promise<void> => {
      return new Promise((resolve, reject) => {
        const existing = document.querySelector(`script[src="${src}"]`);
        if (existing) {
          setTimeout(() => resolve(), 100);
          return;
        }

        const script = document.createElement('script');
        script.src = src;
        script.onload = () => {
          setTimeout(() => resolve(), 50);
        };
        script.onerror = () => reject(new Error(`Failed to load ${src}`));
        document.head.appendChild(script);
      });
    };

    const loadAllScripts = async () => {
      try {
        await loadScript('https://d3js.org/d3.v3.min.js');
        await loadScript('https://d3js.org/topojson.v1.min.js');
        await loadScript('https://cdnjs.cloudflare.com/ajax/libs/datamaps/0.5.9/datamaps.world.min.js');

        if (window.d3 && window.topojson && window.Datamap) {
          setIsLoaded(true);
        } else {
          console.error('Scripts loaded but globals not available');
        }
      } catch (error) {
        console.error('Failed to load datamaps scripts:', error);
      }
    };

    loadAllScripts();
  }, []);

  // Initialize map
  useEffect(() => {
    if (!isLoaded || !mapRef.current || !window.Datamap) return;

    const isDark = document.documentElement.classList.contains('dark');

    // Get theme colors from CSS variables
    const computedStyle = getComputedStyle(document.documentElement);
    const getCSSVar = (varName: string) => {
      const value = computedStyle.getPropertyValue(varName).trim();
      return value;
    };

    // Create color palette based on our theme
    // Light mode: Lighter shades for better contrast
    // Dark mode: Slightly darker but still visible
    const colors = {
      defaultFill: isDark ? 'rgb(34, 34, 34)' : 'rgb(242, 242, 243)',  // --border-primary / --bg-hover
      LOW: isDark ? 'rgba(255, 204, 0, 0.3)' : 'rgba(255, 204, 0, 0.4)',  // --accent-end with opacity
      MEDIUM: isDark ? 'rgba(255, 149, 0, 0.5)' : 'rgba(255, 149, 0, 0.6)',  // --accent-mid with opacity
      HIGH: isDark ? 'rgba(255, 149, 0, 0.7)' : 'rgba(255, 149, 0, 0.8)',  // --accent-mid with higher opacity
      MAJOR: isDark ? 'rgba(255, 59, 48, 0.8)' : 'rgba(255, 59, 48, 0.9)',  // --accent-start with high opacity
      highlight: isDark ? 'rgba(255, 59, 48, 1)' : 'rgba(255, 59, 48, 1)',  // Full accent color on hover
      border: isDark ? 'rgba(34, 34, 34, 0.4)' : 'rgba(229, 229, 229, 0.8)'  // Semi-transparent borders
    };

    const dataMap = new window.Datamap({
      element: mapRef.current,
      projection: 'mercator',
      responsive: true,
      fills: {
        defaultFill: colors.defaultFill,
        LOW: colors.LOW,
        MEDIUM: colors.MEDIUM,
        HIGH: colors.HIGH,
        MAJOR: colors.MAJOR
      },
      data: visitorData,
      geographyConfig: {
        borderColor: colors.border,
        borderWidth: 0.5,
        highlightFillColor: colors.highlight,
        highlightBorderColor: colors.highlight,
        highlightBorderWidth: 1.5,
        popupTemplate: function(geo: any, data: CountryData) {
          if (!data) return '';

          const growUp = '<svg class="inline-block w-4 h-4 text-green-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 7 13.5 15.5 8.5 10.5 2 17"></polyline><polyline points="16 7 22 7 22 13"></polyline></svg>';
          const growDown = '<svg class="inline-block w-4 h-4 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 17 13.5 8.5 8.5 13.5 2 7"></polyline><polyline points="16 17 22 17 22 11"></polyline></svg>';

          return `
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl border border-gray-200 dark:border-gray-700 p-3 min-w-[200px]">
              <div class="flex items-center mb-2">
                <span class="text-xl mr-2">${flagMap[data.short] || ''}</span>
                <span class="text-sm font-semibold text-gray-900 dark:text-white">${data.customName || geo.properties.name}</span>
              </div>
              <div class="space-y-1.5">
                <div class="flex items-center justify-between">
                  <span class="text-xs text-gray-500 dark:text-gray-400">Active Visitors</span>
                  <div class="flex items-center gap-1">
                    <span class="text-sm font-bold text-gray-900 dark:text-white">${data.active.value}</span>
                    <span class="text-xs ${data.active.isGrown ? 'text-green-500' : 'text-red-500'} font-medium">${data.active.percent}%</span>
                    ${data.active.isGrown ? growUp : growDown}
                  </div>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-xs text-gray-500 dark:text-gray-400">New Visitors</span>
                  <div class="flex items-center gap-1">
                    <span class="text-sm font-bold text-gray-900 dark:text-white">${data.new.value}</span>
                    <span class="text-xs ${data.new.isGrown ? 'text-green-500' : 'text-red-500'} font-medium">${data.new.percent}%</span>
                    ${data.new.isGrown ? growUp : growDown}
                  </div>
                </div>
              </div>
            </div>
          `;
        }
      }
    });

    mapInstanceRef.current = dataMap;
    setIsInitialized(true);

    // Debounced resize handler
    const handleResize = () => {
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }
      resizeTimeoutRef.current = setTimeout(() => {
        if (mapInstanceRef.current) {
          mapInstanceRef.current.resize();
        }
      }, 150);
    };

    window.addEventListener('resize', handleResize, { passive: true });

    // Theme change listener
    const handleThemeChange = () => {
      // Re-initialize map with new colors
      setTimeout(() => {
        if (mapRef.current && mapInstanceRef.current) {
          mapRef.current.innerHTML = '';
          setIsInitialized(false);
          // Re-initialization will happen in the next render cycle
        }
      }, 100);
    };

    const observer = new MutationObserver(handleThemeChange);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    });

    return () => {
      window.removeEventListener('resize', handleResize);
      observer.disconnect();
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }
      if (mapInstanceRef.current && mapRef.current) {
        mapRef.current.innerHTML = '';
      }
      mapInstanceRef.current = null;
      setIsInitialized(false);
    };
  }, [isLoaded]);

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-display text-sm text-[rgb(var(--text-primary))] mb-1">
            Global Visitor Distribution
          </h3>
          <p className="text-xs text-[rgb(var(--text-secondary))]">
            {mapStats.totalVisitors} visitors from {mapStats.totalCountries} countries
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-[rgb(var(--accent-mid))] animate-pulse"></div>
          <span className="text-xs text-[rgb(var(--text-secondary))]">{mapStats.growthRate} growth</span>
        </div>
      </div>

      <div className="relative bg-[rgb(var(--bg-subtle))] rounded-lg border border-[rgb(var(--border-primary))] overflow-hidden transition-all duration-300 hover:shadow-[var(--shadow-md)]">
        {!isLoaded && (
          <div className="absolute inset-0 flex items-center justify-center bg-[rgb(var(--bg-subtle))] z-10 backdrop-blur-sm">
            <div className="text-center">
              <div className="w-8 h-8 border-2 border-[rgb(var(--accent-mid))] border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
              <p className="text-sm text-[rgb(var(--text-secondary))] font-medium">Loading map...</p>
            </div>
          </div>
        )}
        <div
          ref={mapRef}
          className="transition-all duration-500"
          style={{
            height: '350px',
            width: '100%',
            opacity: isInitialized ? 1 : 0,
            transform: isInitialized ? 'translateY(0)' : 'translateY(10px)',
          }}
        />
      </div>

      <div className="flex items-center justify-between text-xs">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 group cursor-pointer">
            <div className="w-4 h-4 rounded" style={{ backgroundColor: 'rgba(255, 204, 0, 0.4)' }}></div>
            <span className="text-[rgb(var(--text-secondary))] transition-colors group-hover:text-[rgb(var(--text-primary))]">Low</span>
          </div>
          <div className="flex items-center gap-2 group cursor-pointer">
            <div className="w-4 h-4 rounded" style={{ backgroundColor: 'rgba(255, 149, 0, 0.6)' }}></div>
            <span className="text-[rgb(var(--text-secondary))] transition-colors group-hover:text-[rgb(var(--text-primary))]">Medium</span>
          </div>
          <div className="flex items-center gap-2 group cursor-pointer">
            <div className="w-4 h-4 rounded" style={{ backgroundColor: 'rgba(255, 149, 0, 0.8)' }}></div>
            <span className="text-[rgb(var(--text-secondary))] transition-colors group-hover:text-[rgb(var(--text-primary))]">High</span>
          </div>
          <div className="flex items-center gap-2 group cursor-pointer">
            <div className="w-4 h-4 rounded" style={{ backgroundColor: 'rgba(255, 59, 48, 0.9)' }}></div>
            <span className="text-[rgb(var(--text-secondary))] transition-colors group-hover:text-[rgb(var(--text-primary))]">Major</span>
          </div>
        </div>
        <div className="text-[rgb(var(--text-tertiary))] font-mono">
          Top: {mapStats.topCountries.slice(0, 3).join(', ')}
        </div>
      </div>
    </div>
  );
}