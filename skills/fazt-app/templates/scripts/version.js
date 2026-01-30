#!/usr/bin/env node
/**
 * Version Generator for Fazt Apps
 *
 * Generates version.json with build metadata.
 * Run automatically as part of the build process.
 *
 * Output: { version, build, branch, timestamp }
 */

import { readFileSync, writeFileSync, existsSync } from 'fs'
import { execSync } from 'child_process'

// Read version from package.json (or default)
let version = '1.0.0'
try {
  const pkg = JSON.parse(readFileSync('./package.json', 'utf-8'))
  version = pkg.version || version
} catch (e) {
  // No package.json, use default
}

// Get git info
let build = 'dev'
let branch = 'unknown'
try {
  build = execSync('git rev-parse --short HEAD', { encoding: 'utf-8' }).trim()
  branch = execSync('git rev-parse --abbrev-ref HEAD', { encoding: 'utf-8' }).trim()
} catch (e) {
  // Not in a git repo
}

const versionInfo = {
  version,
  build,
  branch,
  timestamp: new Date().toISOString()
}

// Write to public folder (for dev) and root (for build output)
const targets = ['./public/version.json', './version.json']
for (const target of targets) {
  try {
    writeFileSync(target, JSON.stringify(versionInfo, null, 2))
  } catch (e) {
    // Directory might not exist
  }
}

// Also write to dist if it exists (post-build)
if (existsSync('./dist')) {
  writeFileSync('./dist/version.json', JSON.stringify(versionInfo, null, 2))
}

console.log(`📦 v${version} (${build})`)
