// Root App Component
// Provides layout shell with router-view

import { defineComponent } from 'vue'

export const App = defineComponent({
  name: 'App',

  template: `
    <div class="h-screen flex flex-col overflow-hidden">
      <!-- Router renders current page -->
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>
  `
})
