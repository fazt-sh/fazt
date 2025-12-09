import { useEffect, useRef, useState } from 'react';

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

interface VisitorData {
  value: string;
  percent: string;
  isGrown: boolean;
}

interface CountryData {
  active: VisitorData;
  new: VisitorData;
  fillKey: string;
  short: string;
  customName?: string;
}

export function Datamap() {
  const mapRef = useRef<HTMLDivElement>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);
  const mapInstanceRef = useRef<ReturnType<Window['Datamap']['new']> | null>(null);
  const resizeTimeoutRef = useRef<NodeJS.Timeout>();

  // Define data set outside return so it's accessible in the legend
  const dataSet: Record<string, CountryData> = {
    USA: {
      active: { value: '8,934', percent: '12.5', isGrown: true },
      new: { value: '1,234', percent: '8.3', isGrown: true },
      fillKey: 'MAJOR',
      short: 'us',
      customName: 'United States'
    },
    CHN: {
      active: { value: '6,789', percent: '-5.2', isGrown: false },
      new: { value: '2,345', percent: '-8.5', isGrown: false },
      fillKey: 'MAJOR',
      short: 'cn',
      customName: 'China'
    },
    GBR: {
      active: { value: '5,432', percent: '8.3', isGrown: true },
      new: { value: '987', percent: '15.7', isGrown: true },
      fillKey: 'MAJOR',
      short: 'gb',
      customName: 'United Kingdom'
    },
    DEU: {
      active: { value: '3,456', percent: '3.1', isGrown: true },
      new: { value: '765', percent: '11.4', isGrown: true },
      fillKey: 'MAJOR',
      short: 'de',
      customName: 'Germany'
    },
    FRA: {
      active: { value: '2,987', percent: '15.7', isGrown: true },
      new: { value: '543', percent: '22.1', isGrown: true },
      fillKey: 'MEDIUM',
      short: 'fr',
      customName: 'France'
    },
    JPN: {
      active: { value: '4,123', percent: '-5.2', isGrown: false },
      new: { value: '654', percent: '7.2', isGrown: true },
      fillKey: 'MAJOR',
      short: 'jp',
      customName: 'Japan'
    },
    IND: {
      active: { value: '3,789', percent: '28.3', isGrown: true },
      new: { value: '876', percent: '35.6', isGrown: true },
      fillKey: 'MAJOR',
      short: 'in',
      customName: 'India'
    },
    BRA: {
      active: { value: '1,654', percent: '14.2', isGrown: true },
      new: { value: '321', percent: '18.9', isGrown: true },
      fillKey: 'MEDIUM',
      short: 'br',
      customName: 'Brazil'
    },
    CAN: {
      active: { value: '1,876', percent: '11.4', isGrown: true },
      new: { value: '456', percent: '9.7', isGrown: true },
      fillKey: 'MEDIUM',
      short: 'ca',
      customName: 'Canada'
    },
    AUS: {
      active: { value: '1,234', percent: '7.2', isGrown: true },
      new: { value: '234', percent: '12.8', isGrown: true },
      fillKey: 'MEDIUM',
      short: 'au',
      customName: 'Australia'
    },
    RUS: {
      active: { value: '987', percent: '-12.3', isGrown: false },
      new: { value: '123', percent: '-5.6', isGrown: false },
      fillKey: 'LOW',
      short: 'ru',
      customName: 'Russia'
    },
    MEX: {
      active: { value: '765', percent: '18.9', isGrown: true },
      new: { value: '189', percent: '24.3', isGrown: true },
      fillKey: 'LOW',
      short: 'mx',
      customName: 'Mexico'
    }
  };

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
          // Wait a bit to ensure script is executed
          setTimeout(() => resolve(), 100);
          return;
        }

        const script = document.createElement('script');
        script.src = src;
        script.onload = () => {
          // Wait for script to execute
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

        // Double check everything is loaded
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
      return value || (isDark ? '#374151' : '#d1d5db');
    };

    const dataMap = new window.Datamap({
      element: mapRef.current,
      projection: 'mercator',
      responsive: true,
      fills: {
        defaultFill: getCSSVar('--bg-hover'),
        LOW: getCSSVar('--text-tertiary'),
        MEDIUM: getCSSVar('--accent-mid'),
        MAJOR: getCSSVar('--accent-dark')
      },
      data: dataSet,
      geographyConfig: {
        borderColor: getCSSVar('--border-primary'),
        borderWidth: 0.5,
        highlightFillColor: getCSSVar('--accent-mid'),
        highlightBorderColor: getCSSVar('--accent'),
        highlightBorderWidth: 1,
        popupTemplate: function(geo: any, data: CountryData) {
          if (!data) return '';

          const flagMap: Record<string, string> = {
            us: '🇺🇸', gb: '🇬🇧', cn: '🇨🇳', jp: '🇯🇵',
            fr: '🇫🇷', de: '🇩🇪', in: '🇮🇳', br: '🇧🇷',
            ca: '🇨🇦', au: '🇦🇺', ru: '🇷🇺', mx: '🇲🇽'
          };

          const growUp = '<svg class="inline-block w-4 h-4 text-green-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 7 13.5 15.5 8.5 10.5 2 17"></polyline><polyline points="16 7 22 7 22 13"></polyline></svg>';
          const growDown = '<svg class="inline-block w-4 h-4 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 17 13.5 8.5 8.5 13.5 2 7"></polyline><polyline points="16 17 22 17 22 11"></polyline></svg>';

          return `
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-3 min-w-[200px]">
              <div class="flex mb-2">
                <span class="text-xl mr-2">${flagMap[data.short] || ''}</span>
                <span class="text-sm font-medium text-gray-800 dark:text-white">${data.customName || geo.properties.name}</span>
              </div>
              <div class="flex items-center mb-1">
                <span class="text-xs text-gray-500 dark:text-gray-400">Active:</span>
                <span class="text-sm font-medium text-gray-800 dark:text-white ml-1">${data.active.value}</span>
                <span class="text-xs ${data.active.isGrown ? 'text-green-500' : 'text-red-500'} ml-1">${data.active.percent}%</span>
                ${data.active.isGrown ? growUp : growDown}
              </div>
              <div class="flex items-center">
                <span class="text-xs text-gray-500 dark:text-gray-400">New:</span>
                <span class="text-sm font-medium text-gray-800 dark:text-white ml-1">${data.new.value}</span>
                <span class="text-xs ${data.new.isGrown ? 'text-green-500' : 'text-red-500'} ml-1">${data.new.percent}%</span>
                ${data.new.isGrown ? growUp : growDown}
              </div>
            </div>
          `;
        }
      }
    });

    mapInstanceRef.current = dataMap;
    setIsInitialized(true);

    // Debounced resize handler for better performance
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

    return () => {
      window.removeEventListener('resize', handleResize);
      if (resizeTimeoutRef.current) {
        clearTimeout(resizeTimeoutRef.current);
      }
      // Proper cleanup
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
            Visitors by country for the last 30 days
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-[rgb(var(--accent-mid))] animate-pulse"></div>
          <span className="text-xs text-[rgb(var(--text-secondary))]">Live data</span>
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
            <div className="w-4 h-4 rounded bg-[rgb(var(--text-tertiary))] transition-transform group-hover:scale-110"></div>
            <span className="text-[rgb(var(--text-secondary))] transition-colors group-hover:text-[rgb(var(--text-primary))]">Low</span>
          </div>
          <div className="flex items-center gap-2 group cursor-pointer">
            <div className="w-4 h-4 rounded bg-[rgb(var(--accent-mid))] transition-transform group-hover:scale-110"></div>
            <span className="text-[rgb(var(--text-secondary))] transition-colors group-hover:text-[rgb(var(--text-primary))]">Medium</span>
          </div>
          <div className="flex items-center gap-2 group cursor-pointer">
            <div className="w-4 h-4 rounded bg-[rgb(var(--accent-dark))] transition-transform group-hover:scale-110"></div>
            <span className="text-[rgb(var(--text-secondary))] transition-colors group-hover:text-[rgb(var(--text-primary))]">High</span>
          </div>
        </div>
        <div className="text-[rgb(var(--text-tertiary))] font-mono">
          {Object.keys(dataSet).length} countries tracked
        </div>
      </div>
    </div>
  );
}