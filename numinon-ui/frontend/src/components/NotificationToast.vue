<template>
  <Transition name="slide">
    <div v-if="notifications.length > 0" class="notification-container">
      <div
          v-for="notification in notifications"
          :key="notification.id"
          class="notification"
          :class="notification.type"
      >
        <div class="notification-header">
          <span class="notification-icon">{{ getIcon(notification.type) }}</span>
          <span class="notification-title">{{ notification.title }}</span>
          <button @click="removeNotification(notification.id)" class="close-btn">
            ✕
          </button>
        </div>
        <div v-if="notification.message" class="notification-message">
          {{ notification.message }}
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { EventsOn } from '../../wailsjs/runtime'

interface Notification {
  id: number
  title: string
  message?: string
  type: 'info' | 'success' | 'warning' | 'error'
  timeout?: number
}

const notifications = ref<Notification[]>([])
let nextId = 1

onMounted(() => {
  // Listen for notification events from backend
  EventsOn('notification', (data: { title: string; message?: string; type?: string }) => {
    addNotification(
        data.title,
        data.message,
        (data.type as Notification['type']) || 'info'
    )
  })

  // Listen for app events that might trigger notifications
  EventsOn('app:ready', () => {
    addNotification('Application Ready', 'Numinon C2 Client initialized', 'success')
  })
})

function addNotification(
    title: string,
    message?: string,
    type: Notification['type'] = 'info',
    timeout = 5000
) {
  const notification: Notification = {
    id: nextId++,
    title,
    message,
    type,
    timeout
  }

  notifications.value.push(notification)

  // Auto-remove after timeout
  if (timeout > 0) {
    setTimeout(() => {
      removeNotification(notification.id)
    }, timeout)
  }
}

function removeNotification(id: number) {
  const index = notifications.value.findIndex(n => n.id === id)
  if (index > -1) {
    notifications.value.splice(index, 1)
  }
}

function getIcon(type: Notification['type']): string {
  switch (type) {
    case 'success': return '✅'
    case 'warning': return '⚠️'
    case 'error': return '❌'
    default: return 'ℹ️'
  }
}

// Expose for external use if needed
defineExpose({
  addNotification
})
</script>

<style scoped>
.notification-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-width: 400px;
}

.notification {
  background: #2a2a2a;
  border-radius: 8px;
  padding: 12px 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  border-left: 4px solid #666;
  animation: slideIn 0.3s ease;
}

.notification.success {
  border-left-color: #4ade80;
  background: linear-gradient(to right, rgba(74, 222, 128, 0.1), #2a2a2a);
}

.notification.warning {
  border-left-color: #fbbf24;
  background: linear-gradient(to right, rgba(251, 191, 36, 0.1), #2a2a2a);
}

.notification.error {
  border-left-color: #f87171;
  background: linear-gradient(to right, rgba(248, 113, 113, 0.1), #2a2a2a);
}

.notification.info {
  border-left-color: #60a5fa;
  background: linear-gradient(to right, rgba(96, 165, 250, 0.1), #2a2a2a);
}

.notification-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.notification-icon {
  font-size: 16px;
}

.notification-title {
  flex: 1;
  font-weight: 600;
  color: #fff;
}

.close-btn {
  background: none;
  border: none;
  color: #888;
  cursor: pointer;
  font-size: 18px;
  padding: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.notification-message {
  color: #ccc;
  font-size: 14px;
  line-height: 1.4;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
}

.slide-enter-from {
  transform: translateX(100%);
  opacity: 0;
}

.slide-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
</style>