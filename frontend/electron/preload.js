const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('electronAPI', {
  isElectron: true,
  platform: process.platform,
  linuxDesktop: process.platform === 'linux'
    ? (process.env.XDG_CURRENT_DESKTOP || process.env.DESKTOP_SESSION || '').toLowerCase()
    : null,
  isWayland: process.platform === 'linux' &&
    !!(process.env.WAYLAND_DISPLAY || process.env.XDG_SESSION_TYPE === 'wayland'),
  windowMinimize: () => ipcRenderer.invoke('window-minimize'),
  windowMaximize: () => ipcRenderer.invoke('window-maximize'),
  windowClose: () => ipcRenderer.invoke('window-close'),
  openExternal: (url) => ipcRenderer.invoke('open-external', url),
  federationAuth: (authUrl, callbackOrigin) =>
    ipcRenderer.invoke('federation-auth', authUrl, callbackOrigin),
  getAppVersion: () => ipcRenderer.invoke('get-app-version'),
  checkForUpdates: () => ipcRenderer.invoke('check-for-updates'),
  installUpdate: () => ipcRenderer.invoke('install-update'),
  onUpdateStatus: (callback) => {
    ipcRenderer.on('update-status', (_event, data) => callback(data))
  },
  removeUpdateListener: () => {
    ipcRenderer.removeAllListeners('update-status')
  },
  setPttConfig: (config) => ipcRenderer.invoke('set-ptt-config', config),
  checkAccessibility: () => ipcRenderer.invoke('check-accessibility'),
  requestAccessibility: () => ipcRenderer.invoke('request-accessibility'),
  onPttState: (callback) => {
    ipcRenderer.on('ptt-state', (_event, active) => callback(active))
  },
  onGlobalPttRevoked: (callback) => {
    ipcRenderer.on('global-ptt-revoked', () => callback())
  },
  onGlobalPttAvailable: (callback) => {
    ipcRenderer.on('global-ptt-available', () => callback())
  },
  removePttListener: () => {
    ipcRenderer.removeAllListeners('ptt-state')
    ipcRenderer.removeAllListeners('global-ptt-revoked')
    ipcRenderer.removeAllListeners('global-ptt-available')
  },
})
