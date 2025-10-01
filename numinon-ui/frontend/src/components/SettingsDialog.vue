<template>
  <div v-if="show" class="settings-overlay" @click.self="close">
    <div class="settings-dialog">
      <div class="settings-header">
        <h2>Settings</h2>
        <button class="close-btn" @click="close">✕</button>
      </div>

      <div class="settings-tabs">
        <button
            v-for="tab in tabs"
            :key="tab"
            :class="['tab', { active: activeTab === tab }]"
            @click="activeTab = tab"
        >
          {{ tab }}
        </button>
      </div>

      <div class="settings-content">
        <!-- General Tab -->
        <div v-if="activeTab === 'General'" class="tab-content">
          <div class="setting-group">
            <label>
              <input
                  type="checkbox"
                  v-model="preferences.general.minimizeToTray"
                  @change="updatePreferences"
              />
              Minimize to system tray
            </label>
          </div>

          <div class="setting-group">
            <label>
              <input
                  type="checkbox"
                  v-model="preferences.general.startOnBoot"
                  @change="updatePreferences"
              />
              Start on system boot
            </label>
          </div>

          <div class="setting-group">
            <label>
              <input
                  type="checkbox"
                  v-model="preferences.general.showNotifications"
                  @change="updatePreferences"
              />
              Show desktop notifications
            </label>
          </div>
        </div>

        <!-- Connection Tab -->
        <div v-if="activeTab === 'Connection'" class="tab-content">
          <div class="setting-group">
            <label>Default Server URL</label>
            <input
                type="text"
                v-model="preferences.connection.serverUrl"
                @change="updatePreferences"
                placeholder="ws://localhost:8080/client"
            />
          </div>

          <div class="setting-group">
            <label>
              <input
                  type="checkbox"
                  v-model="preferences.connection.autoConnect"
                  @change="updatePreferences"
              />
              Auto-connect on startup
            </label>
          </div>

          <div class="setting-group">
            <label>Reconnect Delay (seconds)</label>
            <input
                type="number"
                v-model.number="preferences.connection.reconnectDelay"
                @change="updatePreferences"
                min="1"
                max="60"
            />
          </div>
        </div>

        <!-- Theme Tab -->
        <div v-if="activeTab === 'Theme'" class="tab-content">
          <div class="setting-group">
            <label>Theme Mode</label>
            <select v-model="preferences.theme.mode" @change="updatePreferences">
              <option value="dark">Dark</option>
              <option value="light">Light</option>
              <option value="auto">Auto (System)</option>
            </select>
          </div>

          <div class="setting-group">
            <label>Accent Color</label>
            <input
                type="color"
                v-model="preferences.theme.accentColor"
                @change="updatePreferences"
            />
          </div>
        </div>
      </div>

      <div class="settings-footer">
        <div class="footer-left">
          <button class="btn-danger" @click="quitApp">Quit Application</button>
        </div>
        <div class="footer-right">
          <button class="btn-primary" @click="save">Save</button>
          <button class="btn-secondary" @click="close">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { GetPreferences, UpdatePreferences } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime'
import { EventsEmit } from '../../wailsjs/runtime'

const show = ref(false)
const activeTab = ref('General')
const tabs = ['General', 'Connection', 'Theme']

const preferences = reactive({
  general: {
    minimizeToTray: true,
    startOnBoot: false,
    showNotifications: true,
  },
  connection: {
    serverUrl: 'ws://localhost:8080/client',
    autoConnect: false,
    reconnectDelay: 5,
  },
  theme: {
    mode: 'dark',
    accentColor: '#667eea',
  }
})

onMounted(() => {
  // Load preferences on mount
  loadPreferences()

  // Listen for settings show event from tray
  EventsOn('tray:show-settings', () => {
    show.value = true
  })
})

async function loadPreferences() {
  const prefs = await GetPreferences()
  Object.assign(preferences, prefs)
}

async function updatePreferences() {
  // Update preferences in backend
  await UpdatePreferences(preferences)
}

function quitApp() {
  // Emit event to backend to initiate shutdown
  EventsEmit('app:quit-requested')
  show.value = false
}

function save() {
  updatePreferences()
  show.value = false
}

function close() {
  show.value = false
  // Reload preferences to discard changes
  loadPreferences()
}

defineExpose({
  show
})

</script>

<style scoped>
.settings-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.settings-dialog {
  background: #2a2a2a;
  border-radius: 12px;
  width: 600px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #444;
}

.settings-header h2 {
  margin: 0;
  color: #fff;
}

.close-btn {
  background: none;
  border: none;
  color: #888;
  font-size: 24px;
  cursor: pointer;
  padding: 0;
  width: 30px;
  height: 30px;
}

.close-btn:hover {
  color: #fff;
}

.settings-tabs {
  display: flex;
  padding: 0 20px;
  background: #1a1a1a;
  border-bottom: 1px solid #444;
}

.tab {
  background: none;
  border: none;
  color: #888;
  padding: 15px 20px;
  cursor: pointer;
  transition: all 0.3s;
  border-bottom: 2px solid transparent;
}

.tab:hover {
  color: #fff;
}

.tab.active {
  color: #667eea;
  border-bottom-color: #667eea;
}

.settings-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.setting-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.setting-group label {
  color: #ccc;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.setting-group input[type="text"],
.setting-group input[type="number"],
.setting-group select {
  padding: 8px 12px;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 4px;
  color: #fff;
  font-size: 14px;
}

.setting-group input[type="checkbox"] {
  width: 18px;
  height: 18px;
}

.setting-group input[type="color"] {
  width: 60px;
  height: 40px;
  border: 1px solid #444;
  border-radius: 4px;
  background: #1a1a1a;
  cursor: pointer;
}

.settings-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 20px;
  border-top: 1px solid #444;
}

.btn-primary {
  padding: 8px 20px;
  background: #667eea;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-weight: bold;
  cursor: pointer;
}

.btn-primary:hover {
  background: #5568d3;
}

.btn-secondary {
  padding: 8px 20px;
  background: #444;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-secondary:hover {
  background: #555;
}

.settings-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-top: 1px solid #444;
}

.footer-left {
  flex: 1;
}

.footer-right {
  display: flex;
  gap: 10px;
}

.btn-danger {
  padding: 8px 20px;
  background: #f87171;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-weight: bold;
  cursor: pointer;
}

.btn-danger:hover {
  background: #ef4444;
}
</style>