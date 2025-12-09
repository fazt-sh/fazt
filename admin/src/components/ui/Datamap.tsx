import { useEffect, useRef, useState } from 'react';

// We'll import datamaps as a global script
declare global {
  interface Window {
    Datamap: any;
    d3: any;
    topojson: any;
  }
}

interface VisitorData {
  visitors: string;
  change: string;
  isGrown: boolean;
}

interface CountryData {
  active?: VisitorData;
  new?: VisitorData;
  fillKey: string;
  short: string;
  customName?: string;
}

export function Datamap() {
  const mapRef = useRef<HTMLDivElement>(null);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    // Don't reload if already loaded
    if (window.Datamap && window.d3 && window.topojson) {
      setIsLoaded(true);
      return;
    }

    // Load scripts in order: D3 -> TopoJSON -> Datamaps
    const loadScripts = async () => {
      // Load D3.js
      await new Promise<void>((resolve) => {
        const d3Script = document.createElement('script');
        d3Script.src = 'https://cdn.jsdelivr.net/npm/d3@3.5.17/d3.min.js';
        d3Script.onload = () => resolve();
        document.head.appendChild(d3Script);
      });

      // Load TopoJSON
      await new Promise<void>((resolve) => {
        const topoScript = document.createElement('script');
        topoScript.src = 'https://cdn.jsdelivr.net/npm/topojson@1.6.27/build/topojson.min.js';
        topoScript.onload = () => resolve();
        document.head.appendChild(topoScript);
      });

      // Load Datamaps
      await new Promise<void>((resolve) => {
        const datamapScript = document.createElement('script');
        datamapScript.src = 'https://cdn.jsdelivr.net/npm/datamaps@0.5.9/dist/datamaps.world.min.js';
        datamapScript.onload = () => resolve();
        document.head.appendChild(datamapScript);
      });

      setIsLoaded(true);
    };

    loadScripts();

    // Cleanup
    return () => {
      // Don't remove scripts as they might be used elsewhere
    };
  }, []);

  // Initialize map when scripts are loaded
  useEffect(() => {
    if (!isLoaded || !mapRef.current || !window.Datamap) return;

    // Data for the map
    const dataset: Record<string, CountryData> = {
      USA: {
        active: {
          value: '8,934',
          percent: '12.5',
          isGrown: true
        },
        new: {
          value: '1,234',
          percent: '8.3',
          isGrown: true
        },
        fillKey: 'MAJOR',
        short: 'us',
        customName: 'United States'
      },
      CHN: {
        active: {
          value: '6,789',
          percent: '-5.2',
          isGrown: false
        },
        new: {
          value: '2,345',
          percent: '-8.5',
          isGrown: false
        },
        fillKey: 'MAJOR',
        short: 'cn',
        customName: 'China'
      },
      GBR: {
        active: {
          value: '5,432',
          percent: '8.3',
          isGrown: true
        },
        new: {
          value: '987',
          percent: '15.7',
          isGrown: true
        },
        fillKey: 'MAJOR',
        short: 'gb',
        customName: 'United Kingdom'
      },
      DEU: {
        active: {
          value: '3,456',
          percent: '3.1',
          isGrown: true
        },
        new: {
          value: '765',
          percent: '11.4',
          isGrown: true
        },
        fillKey: 'MAJOR',
        short: 'de',
        customName: 'Germany'
      },
      FRA: {
        active: {
          value: '2,987',
          percent: '15.7',
          isGrown: true
        },
        new: {
          value: '543',
          percent: '22.1',
          isGrown: true
        },
        fillKey: 'MEDIUM',
        short: 'fr',
          customName: 'France'
        },
        JPN: {
          active: {
            value: '4,123',
            percent: '-5.2',
            isGrown: false
          },
          new: {
            value: '654',
            percent: '7.2',
            isGrown: true
          },
          fillKey: 'MAJOR',
          short: 'jp',
          customName: 'Japan'
        },
        IND: {
          active: {
            value: '3,789',
            percent: '28.3',
            isGrown: true
          },
          new: {
            value: '876',
            percent: '35.6',
            isGrown: true
          },
          fillKey: 'MAJOR',
          short: 'in',
          customName: 'India'
        },
        BRA: {
          active: {
            value: '1,654',
            percent: '14.2',
            isGrown: true
          },
          new: {
            value: '321',
            percent: '18.9',
            isGrown: true
          },
          fillKey: 'MEDIUM',
          short: 'br',
          customName: 'Brazil'
        },
        CAN: {
          active: {
            value: '1,876',
            percent: '11.4',
            isGrown: true
          },
          new: {
            value: '456',
            percent: '9.7',
            isGrown: true
          },
          fillKey: 'MEDIUM',
          short: 'ca',
          customName: 'Canada'
        },
        AUS: {
          active: {
            value: '1,234',
            percent: '7.2',
            isGrown: true
          },
          new: {
            value: '234',
            percent: '12.8',
            isGrown: true
          },
          fillKey: 'MEDIUM',
          short: 'au',
          customName: 'Australia'
        },
        RUS: {
          active: {
            value: '987',
            percent: '-12.3',
            isGrown: false
          },
          new: {
            value: '123',
            percent: '-5.6',
            isGrown: false
          },
          fillKey: 'LOW',
          short: 'ru',
          customName: 'Russia'
        },
        MEX: {
          active: {
            value: '765',
            percent: '18.9',
            isGrown: true
          },
          new: {
            value: '189',
            percent: '24.3',
            isGrown: true
          },
          fillKey: 'LOW',
          short: 'mx',
          customName: 'Mexico'
        }
      };

      // Check if dark mode is active
      const isDarkMode = document.documentElement.classList.contains('dark');

      // Create the datamap
      const dataMap = new window.Datamap({
        element: mapRef.current,
        projection: 'mercator',
        responsive: true,
        fills: {
          defaultFill: isDarkMode ? '#374151' : '#d1d5db',
          LOW: isDarkMode ? '#6b7280' : '#9ca3af',
          MEDIUM: isDarkMode ? '#3b82f6' : '#60a5fa',
          MAJOR: isDarkMode ? '#1e40af' : '#3b82f6'
        },
        data: dataset,
        geographyConfig: {
          borderColor: isDarkMode ? 'rgba(55, 65, 81, 0.5)' : 'rgba(209, 213, 219, 0.8)',
          borderWidth: 0.5,
          highlightFillColor: 'rgba(139, 92, 246, 0.7)',
          highlightBorderColor: 'rgba(139, 92, 246, 1)',
          highlightBorderWidth: 1,
          popupTemplate: function(geo: any, data: any) {
            if (!data) return null;

            const growUp = '<svg class="inline-block w-4 h-4 text-green-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 7 13.5 15.5 8.5 10.5 2 17"></polyline><polyline points="16 7 22 7 22 13"></polyline></svg>';
            const growDown = '<svg class="inline-block w-4 h-4 text-red-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 17 13.5 8.5 8.5 13.5 2 7"></polyline><polyline points="16 17 22 17 22 11"></polyline></svg>';

            return `
              <div class="p-3 rounded-lg shadow-lg bg-[rgb(var(--bg-elevated))] border border-[rgb(var(--border-primary))] min-w-[200px]">
                <div class="flex mb-2">
                  <div class="me-2">
                    <span class="text-xl">${data.short === 'us' ? '🇺🇸' : data.short === 'gb' ? '🇬🇧' : data.short === 'cn' ? '🇨🇳' : data.short === 'jp' ? '🇯🇵' : data.short === 'fr' ? '🇫🇷' : data.short === 'de' ? '🇩🇪' : data.short === 'in' ? '🇮🇳' : data.short === 'br' ? '🇧🇷' : data.short === 'ca' ? '🇨🇦' : data.short === 'au' ? '🇦🇺' : data.short === 'ru' ? '🇷🇺' : data.short === 'mx' ? '🇲🇽' : ''}</span>
                  </div>
                  <span class="text-sm font-medium text-[rgb(var(--text-primary))]">${data.customName || geo.properties.name}</span>
                </div>
                <div class="flex items-center mb-1">
                  <span class="text-xs text-[rgb(var(--text-secondary))]">Active:</span>
                  &nbsp;<span class="text-sm font-medium text-[rgb(var(--text-primary))]">${data.active.value}</span>
                  &nbsp;<span class="text-xs ${data.active.isGrown ? 'text-green-500' : 'text-red-500'}">${data.active.percent}%</span>
                  &nbsp;${data.active.isGrown ? growUp : growDown}
                </div>
                ${data.new ? `
                  <div class="flex items-center">
                    <span class="text-xs text-[rgb(var(--text-secondary))]">New:</span>
                    &nbsp;<span class="text-sm font-medium text-[rgb(var(--text-primary))]">${data.new.value}</span>
                    &nbsp;<span class="text-xs ${data.new.isGrown ? 'text-green-500' : 'text-red-500'}">${data.new.percent}%</span>
                    &nbsp;${data.new.isGrown ? growUp : growDown}
                  </div>
                ` : ''}
              </div>
            `;
          }
        }
      });

      // Make map responsive
      const handleResize = () => {
        dataMap.resize();
      };
      window.addEventListener('resize', handleResize);

      // Store reference for cleanup
      (mapRef.current as any).mapInstance = dataMap;

      // Cleanup
      return () => {
        window.removeEventListener('resize', handleResize);
        if (mapRef.current && (mapRef.current as any).mapInstance) {
          (mapRef.current as any).mapInstance = null;
        }
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

      {/* Map container */}
      <div className="relative bg-[rgb(var(--bg-subtle))] rounded-lg border border-[rgb(var(--border-primary))] overflow-hidden">
        {!isLoaded && (
          <div className="absolute inset-0 flex items-center justify-center bg-[rgb(var(--bg-subtle))] z-10">
            <div className="text-center">
              <div className="w-8 h-8 border-2 border-[rgb(var(--accent-mid))] border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
              <p className="text-sm text-[rgb(var(--text-secondary))]">Loading map...</p>
            </div>
          </div>
        )}
        <div
          ref={mapRef}
          style={{ height: '350px', width: '100%', opacity: isLoaded ? 1 : 0 }}
        />
      </div>

      {/* Legend */}
      <div className="flex items-center justify-between text-xs">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 rounded bg-gray-400 dark:bg-gray-600"></div>
            <span className="text-[rgb(var(--text-secondary))]">Low activity</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 rounded bg-blue-400 dark:bg-blue-500"></div>
            <span className="text-[rgb(var(--text-secondary))]">Medium activity</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 rounded bg-blue-600 dark:bg-blue-800"></div>
            <span className="text-[rgb(var(--text-secondary))]">High activity</span>
          </div>
        </div>
        <div className="text-[rgb(var(--text-secondary))]">
          14 countries tracked
        </div>
      </div>
    </div>
  );
}