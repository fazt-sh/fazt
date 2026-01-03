# Geo Library

## Summary

Geographic primitives: distance calculations, point-in-polygon, and IP
geolocation. Embedded geodata means location features work offline
with zero API costs.

## The Problem

```javascript
// Today's options for "find nearby stores":
// 1. Google Maps API - $$$, API keys, latency
// 2. PostGIS - Heavy dependency, complex setup
// 3. DIY haversine - Easy to get wrong, no IP lookup

// IP geolocation options:
// 1. MaxMind API - Subscription cost
// 2. Self-host GeoLite2 - 60MB+ database, manual updates
// 3. Third-party APIs - Privacy concerns, latency, costs
```

## The Solution

```javascript
// Distance between two points
const km = fazt.lib.geo.distance(
  40.7128, -74.0060,  // New York
  51.5074, -0.1278    // London
);
// 5570.22

// IP to location (embedded database)
const loc = fazt.lib.geo.fromIP('8.8.8.8');
// { country: 'US', region: 'CA', city: 'Mountain View',
//   lat: 37.386, lon: -122.084, timezone: 'America/Los_Angeles' }

// Point in polygon
const inside = fazt.lib.geo.contains(deliveryZone, userLocation);
```

## Usage

### Distance Calculation

```javascript
// Haversine formula - great-circle distance
const km = fazt.lib.geo.distance(lat1, lon1, lat2, lon2);

// With units
const miles = fazt.lib.geo.distance(lat1, lon1, lat2, lon2, { unit: 'mi' });
const meters = fazt.lib.geo.distance(lat1, lon1, lat2, lon2, { unit: 'm' });

// Example: New York to Los Angeles
fazt.lib.geo.distance(40.7128, -74.0060, 34.0522, -118.2437);
// 3935.75 km
```

### IP Geolocation

```javascript
// From IP address
const loc = await fazt.lib.geo.fromIP('203.0.113.50');
// {
//   ip: '203.0.113.50',
//   country: 'AU',
//   countryName: 'Australia',
//   region: 'NSW',
//   city: 'Sydney',
//   lat: -33.8688,
//   lon: 151.2093,
//   timezone: 'Australia/Sydney',
//   isp: 'Example ISP'
// }

// From current request (in serverless handler)
const loc = await fazt.lib.geo.fromIP(request.ip);

// Country only (faster, smaller dataset)
const country = await fazt.lib.geo.countryFromIP('8.8.8.8');
// { country: 'US', countryName: 'United States' }
```

### Point in Polygon

```javascript
// Define a polygon (delivery zone, geofence, etc.)
const polygon = [
  [40.7484, -73.9857],  // Point 1
  [40.7580, -73.9855],  // Point 2
  [40.7614, -73.9776],  // Point 3
  [40.7505, -73.9764],  // Point 4
  [40.7484, -73.9857]   // Close the polygon
];

const point = [40.7527, -73.9772];

const inside = fazt.lib.geo.contains(polygon, point);
// true

// GeoJSON format also supported
const geojson = {
  type: 'Polygon',
  coordinates: [[...]]
};
fazt.lib.geo.contains(geojson, point);
```

### Bounding Box

```javascript
// Check if point is within bounding box (faster than polygon)
const bbox = {
  minLat: 40.70, maxLat: 40.80,
  minLon: -74.02, maxLon: -73.95
};

fazt.lib.geo.inBounds(bbox, 40.75, -73.99);
// true

// Get bounding box for a set of points
const bounds = fazt.lib.geo.bounds([
  [40.7128, -74.0060],
  [40.7580, -73.9855],
  [40.7484, -73.9857]
]);
// { minLat: 40.7128, maxLat: 40.7580, minLon: -74.0060, maxLon: -73.9855 }
```

### Nearby Search Helper

```javascript
// Find points within radius
const stores = await fazt.storage.ds.find('stores', {});

const nearby = fazt.lib.geo.nearby(
  stores,
  { lat: 40.7128, lon: -74.0060 },  // Center point
  {
    radius: 10,           // 10 km
    latField: 'latitude', // Field names in your data
    lonField: 'longitude',
    limit: 20,
    sort: 'distance'      // Nearest first
  }
);
// [{ ...store, _distance: 1.2 }, { ...store, _distance: 3.4 }, ...]
```

### Timezone from Coordinates

```javascript
// Get timezone for coordinates
const tz = fazt.lib.geo.timezone(40.7128, -74.0060);
// 'America/New_York'

// Useful when you have GPS but user hasn't set timezone
const userTz = fazt.lib.geo.timezone(user.lat, user.lon);
const localTime = fazt.lib.timezone.now(userTz);
```

### Country from Coordinates

```javascript
// Reverse lookup: coordinates to country
const country = fazt.lib.geo.countryAt(48.8566, 2.3522);
// { country: 'FR', countryName: 'France' }
```

## JS API

```javascript
// Distance
fazt.lib.geo.distance(lat1, lon1, lat2, lon2, options?)
// options: { unit: 'km' | 'mi' | 'm' }
// Returns: number

// IP Geolocation
fazt.lib.geo.fromIP(ip)
// Returns: { country, countryName, region, city, lat, lon, timezone, isp }

fazt.lib.geo.countryFromIP(ip)
// Returns: { country, countryName }

// Geometry
fazt.lib.geo.contains(polygon, point)
// Returns: boolean

fazt.lib.geo.inBounds(bbox, lat, lon)
// Returns: boolean

fazt.lib.geo.bounds(points)
// Returns: { minLat, maxLat, minLon, maxLon }

// Lookup
fazt.lib.geo.timezone(lat, lon)
// Returns: string (IANA timezone)

fazt.lib.geo.countryAt(lat, lon)
// Returns: { country, countryName }

// Utilities
fazt.lib.geo.nearby(items, center, options)
// Returns: items with _distance field, sorted
```

## HTTP Endpoint

```
GET /_services/geo/ip?ip=8.8.8.8
GET /_services/geo/ip  (uses request IP)

Response:
{
  "ip": "8.8.8.8",
  "country": "US",
  "countryName": "United States",
  "region": "CA",
  "city": "Mountain View",
  "lat": 37.386,
  "lon": -122.084,
  "timezone": "America/Los_Angeles"
}
```

## Go Libraries

```go
import (
    // IP Geolocation (embedded database)
    "github.com/oschwald/maxminddb-golang"  // ~10KB reader
    // Embedded GeoLite2-City.mmdb          // ~5MB data

    // Geometry (pure Go)
    "github.com/paulmach/orb"               // ~50KB
)

// Haversine is simple math, no library needed
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 6371 // Earth radius in km
    dLat := toRad(lat2 - lat1)
    dLon := toRad(lon2 - lon1)
    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
         math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
         math.Sin(dLon/2)*math.Sin(dLon/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return R * c
}
```

## Embedded Data

| Dataset          | Size   | Coverage             |
| ---------------- | ------ | -------------------- |
| GeoLite2-Country | ~1.5MB | Country from IP      |
| GeoLite2-City    | ~5MB   | City-level from IP   |
| Timezone shapes  | ~500KB | Timezone from coords |
| Country borders  | ~2MB   | Country from coords  |

**Total embedded: ~5-8MB** (configurable at build time)

Default build includes City-level IP + Timezone shapes.

## Common Patterns

### Store Locator

```javascript
// api/stores/nearby.js
module.exports = async (request) => {
  const { lat, lon, radius = 25 } = request.query;

  const stores = await fazt.storage.ds.find('stores', {
    active: true
  });

  const nearby = fazt.lib.geo.nearby(stores, { lat, lon }, {
    radius: parseFloat(radius),
    latField: 'lat',
    lonField: 'lon',
    limit: 20
  });

  return {
    stores: nearby.map(s => ({
      id: s.id,
      name: s.name,
      address: s.address,
      distance: s._distance.toFixed(1) + ' km'
    }))
  };
};
```

### Delivery Zone Check

```javascript
const deliveryZones = await fazt.storage.ds.find('delivery_zones', {});

async function canDeliver(address) {
  // Geocode address to coordinates (via external API or stored)
  const { lat, lon } = address;

  for (const zone of deliveryZones) {
    if (fazt.lib.geo.contains(zone.polygon, [lat, lon])) {
      return { canDeliver: true, zone: zone.name, fee: zone.deliveryFee };
    }
  }

  return { canDeliver: false };
}
```

### Fraud Detection

```javascript
async function checkLoginLocation(userId, requestIP) {
  const user = await fazt.storage.ds.findOne('users', { id: userId });
  const currentLoc = await fazt.lib.geo.fromIP(requestIP);

  if (!user.lastLoginCountry) {
    return { suspicious: false };
  }

  if (currentLoc.country !== user.lastLoginCountry) {
    const lastLogin = user.lastLoginAt;
    const hoursSince = (Date.now() - lastLogin) / (1000 * 60 * 60);

    // Impossible travel: different country in < 3 hours
    if (hoursSince < 3) {
      return {
        suspicious: true,
        reason: 'impossible_travel',
        from: user.lastLoginCountry,
        to: currentLoc.country
      };
    }
  }

  return { suspicious: false };
}
```

### Auto-Detect User Timezone

```javascript
async function getUserTimezone(request, userId) {
  const user = await fazt.storage.ds.findOne('users', { id: userId });

  // User has set preference
  if (user.timezone) {
    return user.timezone;
  }

  // Infer from IP
  const loc = await fazt.lib.geo.fromIP(request.ip);
  return loc.timezone || 'UTC';
}
```

### GDPR/Compliance Check

```javascript
async function checkCompliance(request) {
  const loc = await fazt.lib.geo.fromIP(request.ip);

  const gdprCountries = ['DE', 'FR', 'IT', 'ES', 'NL', ...];
  const isGDPR = gdprCountries.includes(loc.country);

  return {
    requiresCookieConsent: isGDPR,
    requiresDataExport: isGDPR,
    taxRegion: loc.country
  };
}
```

## Data Updates

GeoLite2 database is updated monthly. Fazt releases will include
updated geodata.

For self-hosted instances needing fresher data:
```bash
fazt services geo update  # Download latest GeoLite2
```

## Limits

| Limit              | Default |
| ------------------ | ------- |
| `maxPolygonPoints` | 10,000  |
| `maxNearbyResults` | 1,000   |
| `ipLookupCacheTTL` | 1 hour  |

## Implementation Notes

- ~5-8MB binary addition (embedded geodata)
- Pure Go (no CGO)
- IP lookups are O(log n) with MMDB format
- Haversine distance: microseconds
- Point-in-polygon: O(n) where n = polygon vertices
- LRU cache for IP lookups (reduces repeated queries)

## Accuracy Notes

- IP geolocation is approximate (city-level, not street)
- Accuracy varies: ~95% country, ~75% city
- VPNs/proxies return VPN exit location, not user location
- Mobile IPs may show carrier location, not device location

