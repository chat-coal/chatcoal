const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('electronAPI', {
  isElectron: true,
  platform: process.platform,
  openExternal: (url) => ipcRenderer.invoke('open-external', url),
  federationAuth: (authUrl, callbackOrigin) =>
    ipcRenderer.invoke('federation-auth', authUrl, callbackOrigin),
  getAppVersion: () => ipcRenderer.invoke('get-app-version'),
  checkForUpdates: () => ipcRenderer.invoke('check-for-updates'),
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
