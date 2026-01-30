# Modal Patterns

## Full Modal with Click-Outside Close

**Key**: Both `@click.self` on container AND `@click` on backdrop ensure reliable closing.

```vue
<template>
  <!-- Modal Container -->
  <div
    v-if="showModal"
    class="fixed inset-0 z-50 flex items-end sm:items-center justify-center"
    @click.self="closeModal"
  >
    <!-- Backdrop (blur + dim) -->
    <div
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      @click="closeModal"
    ></div>

    <!-- Modal Content -->
    <div class="relative bg-white dark:bg-neutral-900 w-full sm:w-[480px] sm:rounded-2xl rounded-t-2xl p-6 max-h-[90vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">Modal Title</h2>

      <!-- Content -->
      <div class="space-y-4">
        <!-- Your content here -->
      </div>

      <!-- Actions -->
      <div class="flex gap-3 pt-4">
        <button
          @click="closeModal"
          class="flex-1 py-3 bg-neutral-100 dark:bg-neutral-800 rounded-xl"
        >
          Cancel
        </button>
        <button
          @click="handleSubmit"
          class="flex-1 py-3 bg-blue-500 text-white rounded-xl"
        >
          Confirm
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'

export default {
  setup() {
    const showModal = ref(false)

    function closeModal() {
      showModal.value = false
    }

    function handleSubmit() {
      // Handle action
      closeModal()
    }

    // Keyboard shortcut
    function handleKeydown(e) {
      if (e.key === 'Escape') closeModal()
    }

    onMounted(() => {
      document.addEventListener('keydown', handleKeydown)
    })

    onUnmounted(() => {
      document.removeEventListener('keydown', handleKeydown)
    })

    return { showModal, closeModal, handleSubmit }
  }
}
</script>
```

## Bottom Sheet (Mobile-Optimized)

```vue
<div
  v-if="showSheet"
  class="fixed inset-0 z-50 flex items-end justify-center"
  @click.self="closeSheet"
>
  <div class="absolute inset-0 bg-black/50" @click="closeSheet"></div>

  <!-- Slides up from bottom on mobile, centered on desktop -->
  <div class="relative bg-white dark:bg-neutral-900 w-full sm:w-[480px] rounded-t-2xl sm:rounded-2xl p-6 max-h-[85vh] overflow-y-auto">
    <!-- Drag handle (mobile) -->
    <div class="sm:hidden flex justify-center mb-2">
      <div class="w-12 h-1 bg-neutral-300 dark:bg-neutral-700 rounded-full"></div>
    </div>

    <h2 class="text-xl font-semibold mb-4">Sheet Title</h2>
    <!-- Content -->
  </div>
</div>
```

## Confirmation Dialog

```vue
<div v-if="showConfirm" class="fixed inset-0 z-50 flex items-center justify-center">
  <div class="absolute inset-0 bg-black/50" @click="closeConfirm"></div>

  <div class="relative bg-white dark:bg-neutral-900 w-[320px] rounded-2xl p-6 mx-4">
    <h3 class="text-lg font-semibold mb-2">Are you sure?</h3>
    <p class="text-neutral-600 dark:text-neutral-400 text-sm mb-6">
      This action cannot be undone.
    </p>

    <div class="flex gap-3">
      <button
        @click="closeConfirm"
        class="flex-1 py-2 bg-neutral-100 dark:bg-neutral-800 rounded-lg"
      >
        Cancel
      </button>
      <button
        @click="handleConfirm"
        class="flex-1 py-2 bg-red-500 text-white rounded-lg"
      >
        Delete
      </button>
    </div>
  </div>
</div>
```

## Settings Panel

```vue
<div v-if="showSettings" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center">
  <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showSettings = false"></div>

  <div class="relative bg-white dark:bg-neutral-900 w-full sm:w-[480px] sm:rounded-2xl rounded-t-2xl p-6 max-h-[90vh] overflow-y-auto">
    <h2 class="text-xl font-semibold mb-4">Settings</h2>

    <div class="space-y-6">
      <!-- Theme Selector -->
      <div>
        <label class="block text-sm font-medium mb-2">Theme</label>
        <div class="grid grid-cols-3 gap-2">
          <button
            v-for="theme in ['light', 'dark', 'system']"
            :key="theme"
            @click="updateSetting('theme', theme)"
            :class="settings.theme === theme ? 'bg-blue-500 text-white' : 'bg-neutral-100 dark:bg-neutral-800'"
            class="py-2 rounded-lg font-medium capitalize touch-active"
          >
            {{ theme }}
          </button>
        </div>
      </div>

      <!-- Toggle -->
      <div class="flex items-center justify-between">
        <span>Sound Effects</span>
        <button
          @click="updateSetting('soundEnabled', !settings.soundEnabled)"
          :class="settings.soundEnabled ? 'bg-blue-500' : 'bg-neutral-300 dark:bg-neutral-700'"
          class="w-12 h-7 rounded-full relative touch-active transition-colors"
        >
          <div
            :class="settings.soundEnabled ? 'translate-x-6' : 'translate-x-1'"
            class="absolute top-1 w-5 h-5 bg-white rounded-full transition-transform"
          ></div>
        </button>
      </div>
    </div>
  </div>
</div>
```

## Form Modal

```vue
<div v-if="showForm" class="fixed inset-0 z-50 flex items-end sm:items-center justify-center">
  <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="closeForm"></div>

  <div class="relative bg-white dark:bg-neutral-900 w-full sm:w-[480px] sm:rounded-2xl rounded-t-2xl p-6 max-h-[90vh] overflow-y-auto">
    <h2 class="text-xl font-semibold mb-4">{{ editing ? 'Edit' : 'New' }} Item</h2>

    <form @submit.prevent="handleSubmit" class="space-y-4">
      <div>
        <label class="block text-sm font-medium mb-1">Name</label>
        <input
          v-model="form.name"
          required
          class="w-full px-3 py-2 border border-neutral-200 dark:border-neutral-700 bg-transparent rounded-lg"
          placeholder="Enter name..."
        />
      </div>

      <div>
        <label class="block text-sm font-medium mb-1">Category</label>
        <select
          v-model="form.category"
          required
          class="w-full px-3 py-2 border border-neutral-200 dark:border-neutral-700 bg-white dark:bg-neutral-900 rounded-lg"
        >
          <option v-for="cat in categories" :key="cat.id" :value="cat.id">
            {{ cat.name }}
          </option>
        </select>
      </div>

      <div class="flex gap-3 pt-4">
        <button
          type="button"
          @click="closeForm"
          class="flex-1 py-3 bg-neutral-100 dark:bg-neutral-800 rounded-xl font-medium"
        >
          Cancel
        </button>
        <button
          type="submit"
          class="flex-1 py-3 bg-blue-500 text-white rounded-xl font-medium"
        >
          Save
        </button>
      </div>
    </form>
  </div>
</div>
```

## Key Points

✓ Use `fixed inset-0 z-50` for full-screen overlay
✓ Add **both** `@click.self` and backdrop `@click` for reliable closing
✓ Use `backdrop-blur-sm` for modern blur effect
✓ Animate on mobile: `items-end` (slides up) vs `items-center` (desktop)
✓ Add `Escape` key handler for accessibility
✓ Use `max-h-[90vh] overflow-y-auto` for long content
✓ Responsive width: `w-full sm:w-[480px]`
✓ Mobile corners: `rounded-t-2xl` (top only) vs `sm:rounded-2xl` (all sides)
