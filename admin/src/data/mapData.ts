export interface VisitorData {
  value: string;
  percent: string;
  isGrown: boolean;
}

export interface CountryData {
  active: VisitorData;
  new: VisitorData;
  fillKey: 'LOW' | 'MEDIUM' | 'HIGH' | 'MAJOR';
  short: string;
  customName?: string;
}

// Designed data distribution for better visual hierarchy
// Creates a realistic distribution with clear visual patterns
export const visitorData: Record<string, CountryData> = {
  // Major traffic sources -darkest accent (accent-start)
  USA: {
    active: { value: '24,532', percent: '18.2', isGrown: true },
    new: { value: '3,456', percent: '22.4', isGrown: true },
    fillKey: 'MAJOR',
    short: 'us',
    customName: 'United States'
  },
  CHN: {
    active: { value: '19,876', percent: '-5.2', isGrown: false },
    new: { value: '2,134', percent: '-8.5', isGrown: false },
    fillKey: 'MAJOR',
    short: 'cn',
    customName: 'China'
  },
  IND: {
    active: { value: '16,234', percent: '28.3', isGrown: true },
    new: { value: '2,876', percent: '35.6', isGrown: true },
    fillKey: 'MAJOR',
    short: 'in',
    customName: 'India'
  },
  BRA: {
    active: { value: '14,567', percent: '14.2', isGrown: true },
    new: { value: '1,987', percent: '18.9', isGrown: true },
    fillKey: 'MAJOR',
    short: 'br',
    customName: 'Brazil'
  },

  // High traffic - medium-dark accent
  JPN: {
    active: { value: '11,234', percent: '-2.1', isGrown: false },
    new: { value: '1,543', percent: '7.2', isGrown: true },
    fillKey: 'HIGH',
    short: 'jp',
    customName: 'Japan'
  },
  DEU: {
    active: { value: '9,876', percent: '8.3', isGrown: true },
    new: { value: '1,234', percent: '11.4', isGrown: true },
    fillKey: 'HIGH',
    short: 'de',
    customName: 'Germany'
  },
  GBR: {
    active: { value: '8,765', percent: '12.5', isGrown: true },
    new: { value: '1,098', percent: '15.7', isGrown: true },
    fillKey: 'HIGH',
    short: 'gb',
    customName: 'United Kingdom'
  },
  FRA: {
    active: { value: '7,432', percent: '6.8', isGrown: true },
    new: { value: '987', percent: '9.3', isGrown: true },
    fillKey: 'HIGH',
    short: 'fr',
    customName: 'France'
  },
  CAN: {
    active: { value: '6,234', percent: '11.4', isGrown: true },
    new: { value: '876', percent: '14.7', isGrown: true },
    fillKey: 'HIGH',
    short: 'ca',
    customName: 'Canada'
  },

  // Medium traffic - medium accent
  AUS: {
    active: { value: '4,123', percent: '9.2', isGrown: true },
    new: { value: '543', percent: '12.8', isGrown: true },
    fillKey: 'MEDIUM',
    short: 'au',
    customName: 'Australia'
  },
  RUS: {
    active: { value: '3,876', percent: '-8.3', isGrown: false },
    new: { value: '432', percent: '-5.6', isGrown: false },
    fillKey: 'MEDIUM',
    short: 'ru',
    customName: 'Russia'
  },
  MEX: {
    active: { value: '2,987', percent: '15.9', isGrown: true },
    new: { value: '387', percent: '22.3', isGrown: true },
    fillKey: 'MEDIUM',
    short: 'mx',
    customName: 'Mexico'
  },
  ESP: {
    active: { value: '2,456', percent: '7.8', isGrown: true },
    new: { value: '298', percent: '11.2', isGrown: true },
    fillKey: 'MEDIUM',
    short: 'es',
    customName: 'Spain'
  },
  KOR: {
    active: { value: '2,134', percent: '4.5', isGrown: true },
    new: { value: '267', percent: '8.9', isGrown: true },
    fillKey: 'MEDIUM',
    short: 'kr',
    customName: 'South Korea'
  },

  // Lower traffic - lighter accent
  ITA: {
    active: { value: '1,765', percent: '3.2', isGrown: true },
    new: { value: '198', percent: '6.7', isGrown: true },
    fillKey: 'LOW',
    short: 'it',
    customName: 'Italy'
  },
  NLD: {
    active: { value: '1,234', percent: '5.6', isGrown: true },
    new: { value: '156', percent: '9.4', isGrown: true },
    fillKey: 'LOW',
    short: 'nl',
    customName: 'Netherlands'
  },
  SWE: {
    active: { value: '987', percent: '2.3', isGrown: true },
    new: { value: '123', percent: '5.8', isGrown: true },
    fillKey: 'LOW',
    short: 'se',
    customName: 'Sweden'
  },
  SGP: {
    active: { value: '876', percent: '8.9', isGrown: true },
    new: { value: '109', percent: '13.4', isGrown: true },
    fillKey: 'LOW',
    short: 'sg',
    customName: 'Singapore'
  },
  ZAF: {
    active: { value: '654', percent: '-3.4', isGrown: false },
    new: { value: '87', percent: '-1.2', isGrown: false },
    fillKey: 'LOW',
    short: 'za',
    customName: 'South Africa'
  }
};

// Helper to get flag emojis
export const flagMap: Record<string, string> = {
  us: '🇺🇸', gb: '🇬🇧', cn: '🇨🇳', jp: '🇯🇵',
  fr: '🇫🇷', de: '🇩🇪', in: '🇮🇳', br: '🇧🇷',
  ca: '🇨🇦', au: '🇦🇺', ru: '🇷🇺', mx: '🇲🇽',
  es: '🇪🇸', kr: '🇰🇷', it: '🇮🇹', nl: '🇳🇱',
  se: '🇸🇪', sg: '🇸🇬', za: '🇿🇦'
};

// Statistics for display
export const mapStats = {
  totalCountries: Object.keys(visitorData).length,
  totalVisitors: '125,432',
  growthRate: '+12.4%',
  topCountries: ['USA', 'CHN', 'IND', 'BRA', 'JPN']
};