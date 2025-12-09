import { useState, useEffect, useRef } from 'react';
import { Card, CardBody } from './Card';
import { Badge } from './Badge';
import { useMockMode } from '../../context/MockContext';
import { mockData } from '../../lib/mockData';
import {
  Server,
  Cpu,
  HardDrive,
  MemoryStick,
  Shield,
  Globe,
  Wifi,
  Activity
} from 'lucide-react';

interface SystemMetricProps {
  title: string;
  value: number;
  max: number;
  unit?: string;
  color: string;
  icon: React.ComponentType<{ className?: string }>;
  format?: 'percentage' | 'bytes' | 'number';
}

// Smart formatting function for different data types
function formatValue(value: number, format?: string, unit?: string): { display: string; sub: string } {
  switch (format) {
    case 'percentage':
      return {
        display: value.toFixed(0),
        sub: '%'
      };

    case 'bytes':
      const units = ['B', 'KB', 'MB', 'GB', 'TB'];
      let bytes = value;
      let unitIndex = 0;

      while (bytes >= 1024 && unitIndex < units.length - 1) {
        bytes /= 1024;
        unitIndex++;
      }

      return {
        display: bytes >= 10 ? bytes.toFixed(0) : bytes.toFixed(1),
        sub: units[unitIndex]
      };

    case 'number':
      if (value >= 1000000) {
        return {
          display: (value / 1000000).toFixed(1),
          sub: 'M'
        };
      } else if (value >= 1000) {
        return {
          display: (value / 1000).toFixed(1),
          sub: 'K'
        };
      }
      return {
        display: value.toFixed(0),
        sub: ''
      };

    default:
      return {
        display: value.toFixed(0),
        sub: unit || ''
      };
  }
}

function RadialGauge({ title, value, max, unit, color, icon: Icon, format = 'number' }: SystemMetricProps) {
  const percentage = Math.min((value / max) * 100, 100);
  const strokeDasharray = 2 * Math.PI * 45; // radius of 45
  const strokeDashoffset = strokeDasharray - (percentage / 100) * strokeDasharray;
  const { display, sub } = formatValue(value, format, unit);

  return (
    <div className="text-center">
      <div className="flex items-center justify-center mb-2">
        <Icon className="h-4 w-4 text-[rgb(var(--text-secondary))]" />
      </div>
      <div className="relative inline-flex items-center justify-center">
        <svg className="transform -rotate-90 w-28 h-28">
          <circle
            cx="56"
            cy="56"
            r="40"
            stroke="rgb(var(--bg-subtle))"
            strokeWidth="10"
            fill="none"
          />
          <circle
            cx="56"
            cy="56"
            r="40"
            stroke={color}
            strokeWidth="10"
            fill="none"
            strokeDasharray={strokeDasharray}
            strokeDashoffset={strokeDashoffset}
            strokeLinecap="round"
            className="transition-all duration-1000 ease-out"
            style={{
              filter: 'drop-shadow(0 0 4px ' + color + '40)'
            }}
          />
        </svg>
        <div className="absolute flex flex-col items-center leading-tight">
          <span className="text-lg font-bold text-[rgb(var(--text-primary))] tabular-nums">
            {display}
          </span>
          {sub && (
            <span className="text-xs font-medium text-[rgb(var(--text-secondary))] uppercase tracking-wide">
              {sub}
            </span>
          )}
        </div>
      </div>
      <p className="text-xs font-medium text-[rgb(var(--text-primary))] mt-2">{title}</p>
    </div>
  );
}

interface InfoCardProps {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string;
  badge?: {
    text: string;
    variant: 'success' | 'warning' | 'error' | 'info';
  };
}

function InfoCard({ icon: Icon, label, value, badge }: InfoCardProps) {
  return (
    <div className="bg-[rgb(var(--bg-subtle))] rounded-lg p-4 hover:bg-[rgb(var(--accent-glow))] transition-all duration-300">
      <div className="flex items-center justify-between mb-2">
        <Icon className="h-4 w-4 text-[rgb(var(--text-secondary))]" />
        {badge && (
          <Badge variant={badge.variant} className="text-xs">
            {badge.text}
          </Badge>
        )}
      </div>
      <p className="text-xs text-[rgb(var(--text-tertiary))] mb-1">{label}</p>
      <p className="text-sm font-semibold text-[rgb(var(--text-primary)] font-mono">{value}</p>
    </div>
  );
}

export function SystemInfo() {
  const { enabled: mockMode } = useMockMode();

  // Mock system data
  const systemInfo = mockMode ? {
    machineSize: 'Large',
    cpu: {
      cores: 4,
      model: 'Intel Xeon E5-2676 v3'
    },
    ram: {
      total: 16,
      used: 8.7,
      utilisation: 54
    },
    storage: {
      total: 500,
      used: 187,
      free: 313
    },
    domain: 'fazt.example.com',
    ip: '192.168.1.100',
    https: {
      enabled: true,
      issuer: 'Let\'s Encrypt',
      expires: '2025-03-09'
    },
    uptime: '14 days, 7 hours',
    region: 'US-East (N. Virginia)'
  } : null;

  if (!systemInfo) {
    return null;
  }

  return (
    <Card variant="bordered" className="p-6"
          style={{
            animation: 'slideIn 0.4s ease-out 0.6s backwards',
          }}>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="font-display text-lg text-[rgb(var(--text-primary))] flex items-center gap-2">
            <Server className="h-5 w-5" />
            System Information
          </h2>
          <p className="text-sm text-[rgb(var(--text-secondary))] mt-1">
            {systemInfo.machineSize} instance • {systemInfo.region}
          </p>
        </div>
        <Badge variant="success" className="gap-1">
          <Activity className="h-3 w-3" />
          Online
        </Badge>
      </div>

      {/* Metrics Grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <RadialGauge
          title="CPU"
          value={35}
          max={100}
          color="rgb(var(--accent-mid))"
          icon={Cpu}
          format="percentage"
        />
        <RadialGauge
          title="RAM"
          value={systemInfo.ram.utilisation}
          max={100}
          color="rgb(var(--accent-start))"
          icon={MemoryStick}
          format="percentage"
        />
        <RadialGauge
          title="Storage"
          value={systemInfo.storage.used * 1024} // Convert to MB for smart formatting
          max={systemInfo.storage.total * 1024}
          color="rgb(var(--accent-end))"
          icon={HardDrive}
          format="bytes"
        />
        <div className="flex flex-col items-center justify-center">
          <div className="flex items-center justify-center mb-2">
            <Activity className="h-4 w-4 text-[rgb(var(--text-secondary))]" />
          </div>
          <div className="text-2xl font-bold text-[rgb(var(--text-primary))] mb-1">
            {systemInfo.uptime.split(',')[0]}
          </div>
          <p className="text-xs font-medium text-[rgb(var(--text-primary))]">Uptime</p>
        </div>
      </div>

      {/* Info Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <InfoCard
          icon={Globe}
          label="Domain"
          value={systemInfo.domain}
        />
        <InfoCard
          icon={Wifi}
          label="IP Address"
          value={systemInfo.ip}
        />
        <InfoCard
          icon={Shield}
          label="HTTPS"
          value={systemInfo.https.issuer}
          badge={{
            text: 'Valid',
            variant: 'success'
          }}
        />
        <InfoCard
          icon={Server}
          label="CPU"
          value={`${systemInfo.cpu.cores} cores`}
        />
      </div>

      {/* Storage Details */}
      <div className="mt-4 p-3 bg-[rgb(var(--bg-subtle))] rounded-lg">
        <div className="flex items-center justify-between text-sm">
          <span className="text-[rgb(var(--text-secondary))]">Storage Usage</span>
          <span className="text-[rgb(var(--text-primary))] font-medium tabular-nums">
            {formatValue(systemInfo.storage.used * 1024, 'bytes').display} {formatValue(systemInfo.storage.used * 1024, 'bytes').sub} / {formatValue(systemInfo.storage.total * 1024, 'bytes').display} {formatValue(systemInfo.storage.total * 1024, 'bytes').sub}
          </span>
        </div>
        <div className="mt-2 h-2 bg-[rgb(var(--bg-primary))] rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-[rgb(var(--accent-start))] to-[rgb(var(--accent-end))] transition-all duration-500"
            style={{ width: `${(systemInfo.storage.used / systemInfo.storage.total) * 100}%` }}
          ></div>
        </div>
        <p className="mt-2 text-xs text-[rgb(var(--text-tertiary))]">
          {formatValue(systemInfo.storage.free * 1024, 'bytes').display} {formatValue(systemInfo.storage.free * 1024, 'bytes').sub} free ({Math.round((systemInfo.storage.free / systemInfo.storage.total) * 100)}%)
        </p>
      </div>
    </Card>
  );
}