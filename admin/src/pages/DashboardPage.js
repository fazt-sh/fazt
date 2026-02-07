import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useHealthStore } from '../stores/health.js'
import { useAppsStore } from '../stores/apps.js'
import { useAliasesStore } from '../stores/aliases.js'
import { useLogsStore } from '../stores/logs.js'
import { useIcons } from '../lib/useIcons.js'
import { usePanel } from '../lib/usePanel.js'
import { formatBytes, formatUptime } from '../lib/format.js'
import FPanel from '../components/FPanel.js'
import StatCard from '../components/StatCard.js'

export default {
  name: 'DashboardPage',
  components: { FPanel, StatCard },
  setup() {
    useIcons()
    const router = useRouter()
    const healthStore = useHealthStore()
    const appsStore = useAppsStore()
    const aliasesStore = useAliasesStore()
    const logsStore = useLogsStore()
    const panel = usePanel('dashboard.system.collapsed', false)

    const stats = computed(() => [
      {
        label: 'Status', icon: 'heart-pulse',
        value: healthStore.data?.status || '-',
        valueClass: healthStore.data?.status === 'healthy' ? 'text-success' : healthStore.data?.status === 'warning' ? 'text-warning' : 'text-error',
        subtitle: healthStore.data?.version ? 'v' + healthStore.data.version : '-',
      },
      {
        label: 'Uptime', icon: 'clock',
        value: formatUptime(healthStore.data?.uptime_seconds || 0),
        subtitle: healthStore.data?.mode || '-',
      },
      {
        label: 'Memory', icon: 'cpu',
        value: (healthStore.data?.memory?.used_mb || healthStore.data?.memory_used_mb || 0).toFixed(1) + ' MB',
        subtitle: (healthStore.data?.runtime?.goroutines || 0) + ' goroutines',
      },
      {
        label: 'Storage', icon: 'database',
        value: formatBytes(appsStore.items.reduce((sum, app) => sum + (app.size_bytes || 0), 0)),
        subtitle: appsStore.items.length + ' apps',
      },
      {
        label: 'Apps', icon: 'layers',
        value: String(appsStore.items.length),
        subtitle: 'deployed',
        clickable: true, route: '/apps',
      },
      {
        label: 'Aliases', icon: 'link',
        value: String(aliasesStore.items.length),
        subtitle: 'configured',
        clickable: true, route: '/aliases',
      },
      {
        label: 'Activity', icon: 'activity',
        value: String(logsStore.stats?.total_count || 0),
        subtitle: 'events logged',
        clickable: true, route: '/logs',
      },
    ])

    const navigateTo = (path) => router.push(path)

    return { panel, stats, navigateTo }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">
          <FPanel title="Dashboard" mode="content"
                  :collapsed="panel.collapsed" @update:collapsed="panel.toggle">
            <template #header-actions>
              <span class="text-caption text-faint ml-auto hide-mobile">7 metrics</span>
            </template>
            <div class="panel-grid grid-4">
              <StatCard v-for="stat in stats" :key="stat.label"
                        :label="stat.label" :icon="stat.icon" :value="stat.value"
                        :value-class="stat.valueClass || 'text-primary'"
                        :subtitle="stat.subtitle"
                        :clickable="!!stat.clickable"
                        @click="stat.route && navigateTo(stat.route)" />
            </div>
          </FPanel>
        </div>
      </div>
    </div>
  `
}
