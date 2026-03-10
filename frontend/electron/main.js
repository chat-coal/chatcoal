import { app, BrowserWindow, session, ipcMain, shell, systemPreferences } from 'electron'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import http from 'node:http'
import fs from 'node:fs'
import pkg from 'electron-updater'
const { autoUpdater } = pkg
import { isWayland, initPortalPtt, stopPortalPtt } from './portal-ptt.js'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const isDev = !app.isPackaged

let mainWindow
let localServer = null

const MIME_TYPES = {
  '.html': 'text/html',
  '.js': 'text/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.webp': 'image/webp',
}

// Serve the production build from a local HTTP server so the app has a proper
// http://localhost origin. This is required for Firebase signInWithPopup — the
// auth handler popup needs to postMessage back to an HTTP(S) origin, not file://.
function startLocalServer() {
  const distPath = path.join(__dirname, '..', 'dist')
  return new Promise((resolve) => {
    const server = http.createServer((req, res) => {
      const urlPath = req.url.split('?')[0]
      const filePath = path.join(distPath, urlPath === '/' ? 'index.html' : urlPath)

      fs.readFile(filePath, (err, data) => {
        if (err) {
          // SPA fallback
          fs.readFile(path.join(distPath, 'index.html'), (_, html) => {
            res.writeHead(200, { 'Content-Type': 'text/html' })
            res.end(html)
          })
          return
        }
        const ext = path.extname(filePath).toLowerCase()
        res.writeHead(200, { 'Content-Type': MIME_TYPES[ext] || 'application/octet-stream' })
        res.end(data)
      })
    })
    server.listen(19532, 'localhost', () => resolve(server))
  })
}

// Global push-to-talk state
let pttKeyCode = null
let pttWebKeyCode = null
let pttEnabled = false
let pttPressed = false
let globalPttAvailable = false
let pttBackend = null // 'portal' | 'uiohook' | null
let uIOhookInstance = null
let accessibilityPollInterval = null

function stopGlobalPtt() {
  if (pttBackend === 'portal') {
    stopPortalPtt()
  }
  if (uIOhookInstance) {
    try { uIOhookInstance.stop() } catch {}
    uIOhookInstance = null
  }
  pttBackend = null
  globalPttAvailable = false
  const canSend = mainWindow && !mainWindow.isDestroyed()
  // If PTT key was held when permission was revoked, release it
  if (pttPressed) {
    pttPressed = false
    if (canSend) mainWindow.webContents.send('ptt-state', false)
  }
  if (canSend) mainWindow.webContents.send('global-ptt-revoked')
}

// Start or stop the accessibility polling interval based on whether PTT
// is enabled. The single interval handles both directions:
//   - Hook running + permission revoked → stop hook, notify renderer
//   - Hook not running + permission granted → start hook, notify renderer
function updateAccessibilityPoll() {
  if (process.platform !== 'darwin') return

  if (pttEnabled && !accessibilityPollInterval) {
    accessibilityPollInterval = setInterval(async () => {
      const trusted = systemPreferences.isTrustedAccessibilityClient(false)
      if (globalPttAvailable && !trusted) {
        console.warn('Accessibility permission revoked — stopping global PTT hook')
        stopGlobalPtt()
      } else if (!globalPttAvailable && trusted) {
        console.log('Accessibility permission granted — starting global PTT hook')
        await initGlobalPtt()
        // Notify renderer that permission is granted (and hook started if
        // uiohook succeeded). Send the event even if the hook didn't fully
        // start — the renderer uses this to hide the permission banner.
        if (mainWindow && !mainWindow.isDestroyed()) {
          mainWindow.webContents.send('global-ptt-available')
        }
      }
    }, 2000)
  } else if (!pttEnabled && accessibilityPollInterval) {
    clearInterval(accessibilityPollInterval)
    accessibilityPollInterval = null
  }
}

async function initGlobalPtt() {
  // On Linux/Wayland, try the XDG GlobalShortcuts portal first.
  // This is the compositor-native approach and doesn't need /dev/input access.
  if (isWayland() && pttWebKeyCode) {
    const ok = await initPortalPtt({
      webKeyCode: pttWebKeyCode,
      onPress() {
        if (!pttEnabled || pttPressed) return
        pttPressed = true
        mainWindow?.webContents.send('ptt-state', true)
      },
      onRelease() {
        pttPressed = false
        if (!pttEnabled) return
        mainWindow?.webContents.send('ptt-state', false)
      },
      onDisconnect() {
        // Portal lost (e.g. compositor restart) — notify renderer to fall
        // back to browser-level PTT until the portal is re-established.
        pttBackend = null
        globalPttAvailable = false
        const canSend = mainWindow && !mainWindow.isDestroyed()
        if (pttPressed) {
          pttPressed = false
          if (canSend) mainWindow.webContents.send('ptt-state', false)
        }
        if (canSend) mainWindow.webContents.send('global-ptt-revoked')
      },
    })
    if (ok) {
      pttBackend = 'portal'
      globalPttAvailable = true
      app.on('will-quit', () => {
        stopGlobalPtt()
        if (accessibilityPollInterval) {
          clearInterval(accessibilityPollInterval)
          accessibilityPollInterval = null
        }
      })
      return
    }
    console.log('Portal unavailable, falling back to uiohook-napi')
  }

  // On macOS, uiohook crashes the process if accessibility is not granted.
  // Check without prompting (false) — the prompt variant (true) can freeze
  // the main process event loop on macOS. If not trusted, the renderer
  // falls back to window-scoped keyboard events.
  if (process.platform === 'darwin') {
    const trusted = systemPreferences.isTrustedAccessibilityClient(false)
    if (!trusted) {
      console.warn('Global PTT unavailable: accessibility permission not granted')
      return
    }
  }

  try {
    const { uIOhook } = await import('uiohook-napi')

    uIOhook.on('keydown', (e) => {
      if (!pttEnabled || e.keycode !== pttKeyCode || pttPressed) return
      pttPressed = true
      mainWindow?.webContents.send('ptt-state', true)
    })

    uIOhook.on('keyup', (e) => {
      if (e.keycode !== pttKeyCode) return
      pttPressed = false
      if (!pttEnabled) return
      mainWindow?.webContents.send('ptt-state', false)
    })

    uIOhook.start()
    uIOhookInstance = uIOhook
    pttBackend = 'uiohook'
    globalPttAvailable = true

    app.on('will-quit', () => {
      stopGlobalPtt()
      if (accessibilityPollInterval) {
        clearInterval(accessibilityPollInterval)
        accessibilityPollInterval = null
      }
    })
  } catch (err) {
    console.warn('Global keyboard hook unavailable:', err.message)
  }
}

async function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 480,
    minHeight: 400,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
    ...(process.platform === 'darwin'
      ? { titleBarStyle: 'hiddenInset', trafficLightPosition: { x: 12, y: 12 } }
      : { frame: false }),
    show: false,
  })

  // Grant microphone permissions automatically (needed for voice chat)
  session.defaultSession.setPermissionRequestHandler((_webContents, permission, callback) => {
    const allowed = ['media', 'mediaKeySystem', 'notifications']
    callback(allowed.includes(permission))
  })

  // Strip Cross-Origin-Opener-Policy from Firebase auth handler responses.
  // COOP: same-origin prevents the auth popup from calling window.close().
  if (!isDev) {
    session.defaultSession.webRequest.onHeadersReceived((details, callback) => {
      if (!details.url.includes('/__/auth/')) return callback({ responseHeaders: details.responseHeaders })
      const responseHeaders = {}
      for (const [key, value] of Object.entries(details.responseHeaders || {})) {
        if (key.toLowerCase() === 'cross-origin-opener-policy') continue
        responseHeaders[key] = value
      }
      callback({ responseHeaders })
    })
  }

  // Handle window.open() calls from the renderer
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    // Allow Firebase auth popups to open as in-app child windows
    // (signInWithPopup / linkWithPopup open /__/auth/handler via window.open)
    if (url.includes('/__/auth/handler')) {
      return {
        action: 'allow',
        overrideBrowserWindowOptions: {
          width: 500,
          height: 700,
          parent: mainWindow,
          modal: true,
          autoHideMenuBar: true,
          webPreferences: {
            contextIsolation: true,
            nodeIntegration: false,
          },
        },
      }
    }

    // Open all other external links in the system browser
    if (url.startsWith('http://') || url.startsWith('https://')) {
      shell.openExternal(url)
    }
    return { action: 'deny' }
  })

  mainWindow.once('ready-to-show', () => {
    mainWindow.show()
    // Check for updates shortly after launch (non-blocking)
    if (!isDev) {
      setTimeout(() => autoUpdater.checkForUpdates().catch(() => {}), 3000)
    }
  })

  // Cmd+Shift+I to toggle DevTools (works in production too)
  mainWindow.webContents.on('before-input-event', (_event, input) => {
    if (input.meta && input.shift && input.key === 'I') {
      mainWindow.webContents.toggleDevTools()
    }
  })

  if (isDev) {
    loadDevServer()
  } else {
    localServer = await startLocalServer()
    const port = localServer.address().port
    mainWindow.loadURL(`http://localhost:${port}`)
  }
}

async function loadDevServer() {
  const url = 'http://localhost:5173'
  // Wait for Vite dev server to be ready
  const maxAttempts = 30
  for (let i = 0; i < maxAttempts; i++) {
    try {
      await fetch(url)
      break
    } catch {
      if (i === maxAttempts - 1) {
        console.error('Vite dev server not reachable at', url)
        app.quit()
        return
      }
      await new Promise((r) => setTimeout(r, 1000))
    }
  }
  mainWindow.loadURL(url)
}

// IPC: Window controls (frameless Windows)
ipcMain.handle('window-minimize', () => mainWindow?.minimize())
ipcMain.handle('window-maximize', () => {
  if (mainWindow?.isMaximized()) mainWindow.unmaximize()
  else mainWindow?.maximize()
})
ipcMain.handle('window-close', () => mainWindow?.close())

// IPC: Open URL in system browser
ipcMain.handle('open-external', (_event, url) => {
  shell.openExternal(url)
})

// IPC: Federation auth — opens auth URL in a child window, intercepts callback
ipcMain.handle('federation-auth', async (_event, authUrl, callbackOrigin) => {
  return new Promise((resolve, reject) => {
    const authWindow = new BrowserWindow({
      width: 600,
      height: 700,
      parent: mainWindow,
      modal: true,
      webPreferences: {
        contextIsolation: true,
        nodeIntegration: false,
      },
    })

    // Intercept navigation to the callback URL
    authWindow.webContents.on('will-navigate', (_e, navUrl) => {
      if (navUrl.startsWith(callbackOrigin + '/federation/callback')) {
        const url = new URL(navUrl)
        const token = url.searchParams.get('token')
        authWindow.close()
        if (token) {
          resolve(token)
        } else {
          reject(new Error('No token in federation callback'))
        }
      }
    })

    // Also intercept redirects (some auth flows use 302)
    authWindow.webContents.on('will-redirect', (_e, navUrl) => {
      if (navUrl.startsWith(callbackOrigin + '/federation/callback')) {
        const url = new URL(navUrl)
        const token = url.searchParams.get('token')
        authWindow.close()
        if (token) {
          resolve(token)
        } else {
          reject(new Error('No token in federation callback'))
        }
      }
    })

    authWindow.on('closed', () => {
      reject(new Error('Auth window closed by user'))
    })

    authWindow.loadURL(authUrl)
  })
})

// IPC: Get app version
ipcMain.handle('get-app-version', () => app.getVersion())

// IPC: Configure global push-to-talk.
// Lazily initialises the global keyboard hook the first time PTT is enabled.
ipcMain.handle('set-ptt-config', async (_event, { keyCode, webKeyCode, enabled }) => {
  const keyChanged = pttWebKeyCode !== webKeyCode
  pttKeyCode = keyCode
  pttWebKeyCode = webKeyCode
  pttEnabled = enabled
  if (!enabled) pttPressed = false

  // KDE persists shortcut bindings globally by ID — tear down the portal
  // session whenever the key changes so a fresh session picks up the new key.
  if (keyChanged && pttBackend === 'portal') {
    stopGlobalPtt()
  }

  // Start the global hook (first time, or after teardown above)
  if (enabled && !globalPttAvailable) {
    await initGlobalPtt()
  }

  // Start/stop the accessibility poll based on PTT state.
  // When PTT is active, the poll auto-starts the hook if permission is
  // granted mid-call, and stops it if permission is revoked.
  updateAccessibilityPoll()

  return globalPttAvailable
})

// IPC: Check if accessibility is granted (no prompt) and start hook if so.
// Returns true when the OS-level permission is confirmed — the hook itself
// may start later (e.g. when the user joins a voice channel).
ipcMain.handle('check-accessibility', async () => {
  if (globalPttAvailable) return true
  if (process.platform === 'darwin') {
    const trusted = systemPreferences.isTrustedAccessibilityClient(false)
    if (!trusted) return false
    // Permission granted — try to start the hook as a side-effect.
    // Even if uiohook fails to start (native module issue, etc.),
    // report the permission as granted so the UI banner hides.
    await initGlobalPtt()
    return true
  }
  await initGlobalPtt()
  return globalPttAvailable
})

// IPC: Prompt for macOS accessibility and retry the global hook.
// Opens System Preferences non-blockingly so the main process never hangs.
ipcMain.handle('request-accessibility', async () => {
  if (globalPttAvailable) return true

  if (process.platform === 'darwin') {
    const trusted = systemPreferences.isTrustedAccessibilityClient(false)
    if (!trusted) {
      shell.openExternal('x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility')
      return false
    }
    await initGlobalPtt()
    return true
  }

  await initGlobalPtt()
  return globalPttAvailable
})

// --- Auto-updater (electron-updater) ---
autoUpdater.autoDownload = true
autoUpdater.autoInstallOnAppQuit = true
autoUpdater.logger = null // silence default logging

function sendUpdateStatus(data) {
  if (mainWindow && !mainWindow.isDestroyed()) {
    mainWindow.webContents.send('update-status', data)
  }
}

autoUpdater.on('update-available', (info) => {
  sendUpdateStatus({ status: 'available', version: info.version })
})

autoUpdater.on('download-progress', (progress) => {
  sendUpdateStatus({ status: 'downloading', percent: Math.round(progress.percent) })
})

autoUpdater.on('update-downloaded', (info) => {
  sendUpdateStatus({ status: 'ready', version: info.version })
})

autoUpdater.on('error', (err) => {
  console.warn('Auto-updater error:', err.message)
  sendUpdateStatus({ status: 'error', message: err.message })
})

ipcMain.handle('check-for-updates', () => {
  if (isDev) return
  autoUpdater.checkForUpdates().catch(() => {})
})

ipcMain.handle('install-update', () => {
  autoUpdater.quitAndInstall(false, true)
})

// Register protocol for deep linking (future use)
if (process.defaultApp) {
  if (process.argv.length >= 2) {
    app.setAsDefaultProtocolClient('chatcoal', process.execPath, [path.resolve(process.argv[1])])
  }
} else {
  app.setAsDefaultProtocolClient('chatcoal')
}

app.whenReady().then(async () => {
  await createWindow()
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    localServer?.close()
    app.quit()
  }
})

app.on('activate', async () => {
  if (BrowserWindow.getAllWindows().length === 0) await createWindow()
})
