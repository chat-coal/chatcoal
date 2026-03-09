// XDG GlobalShortcuts portal for system-wide PTT on Wayland.
// Uses org.freedesktop.portal.GlobalShortcuts to register a push-to-talk
// shortcut that works across all apps without needing /dev/input access.
//
// Portal methods use an async Request/Response pattern:
//   1. Method call returns a Request object path
//   2. Portal emits a Response signal on that path with the actual result
// We intercept Response signals via bus._connection to properly await results.

// Web KeyboardEvent.code → XKB key name for portal preferred_trigger
const CODE_TO_ACCELERATOR = {
  'KeyA': 'a', 'KeyB': 'b', 'KeyC': 'c', 'KeyD': 'd',
  'KeyE': 'e', 'KeyF': 'f', 'KeyG': 'g', 'KeyH': 'h',
  'KeyI': 'i', 'KeyJ': 'j', 'KeyK': 'k', 'KeyL': 'l',
  'KeyM': 'm', 'KeyN': 'n', 'KeyO': 'o', 'KeyP': 'p',
  'KeyQ': 'q', 'KeyR': 'r', 'KeyS': 's', 'KeyT': 't',
  'KeyU': 'u', 'KeyV': 'v', 'KeyW': 'w', 'KeyX': 'x',
  'KeyY': 'y', 'KeyZ': 'z',
  'Digit0': '0', 'Digit1': '1', 'Digit2': '2', 'Digit3': '3',
  'Digit4': '4', 'Digit5': '5', 'Digit6': '6', 'Digit7': '7',
  'Digit8': '8', 'Digit9': '9',
  'Space': 'space', 'Enter': 'Return', 'Tab': 'Tab',
  'Backspace': 'BackSpace', 'Escape': 'Escape', 'CapsLock': 'Caps_Lock',
  'ShiftLeft': 'Shift_L', 'ShiftRight': 'Shift_R',
  'ControlLeft': 'Control_L', 'ControlRight': 'Control_R',
  'AltLeft': 'Alt_L', 'AltRight': 'Alt_R',
  'MetaLeft': 'Super_L', 'MetaRight': 'Super_R',
  'F1': 'F1', 'F2': 'F2', 'F3': 'F3', 'F4': 'F4',
  'F5': 'F5', 'F6': 'F6', 'F7': 'F7', 'F8': 'F8',
  'F9': 'F9', 'F10': 'F10', 'F11': 'F11', 'F12': 'F12',
  'Semicolon': 'semicolon', 'Equal': 'equal', 'Comma': 'comma',
  'Minus': 'minus', 'Period': 'period', 'Slash': 'slash',
  'Backquote': 'grave', 'BracketLeft': 'bracketleft',
  'Backslash': 'backslash', 'BracketRight': 'bracketright',
  'Quote': 'apostrophe',
}

let dbus = null // module reference, lazily imported
let bus = null
let sessionPath = null
let senderToken = null
let signalHandler = null // reference so we can remove it on cleanup
let activeShortcutId = null // current shortcut ID (includes key code)

// Pending Response signal waiters keyed by request object path
const responseWaiters = new Map()
let bindCounter = 0

export function isWayland() {
  return process.platform === 'linux' &&
    !!(process.env.WAYLAND_DISPLAY || process.env.XDG_SESSION_TYPE === 'wayland')
}

/**
 * Subscribe to a portal Response signal BEFORE issuing the method call.
 * Returns a Promise that resolves with the result dict on success.
 */
function waitForResponse(requestPath, timeoutMs = 10000) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      responseWaiters.delete(requestPath)
      reject(new Error('Portal response timed out'))
    }, timeoutMs)

    responseWaiters.set(requestPath, (body) => {
      clearTimeout(timer)
      const [code, results] = body
      if (code === 0) resolve(results)
      else reject(new Error(`Portal request denied (code ${code})`))
    })
  })
}

/**
 * Initialise the GlobalShortcuts portal for PTT.
 * Returns true if the portal is available and the shortcut was bound.
 */
export async function initPortalPtt({ webKeyCode, onPress, onRelease, onDisconnect }) {
  if (!isWayland()) return false

  try {
    dbus = (await import('dbus-next')).default
  } catch {
    console.warn('dbus-next not available, skipping portal PTT')
    return false
  }

  const { Variant, Message, MessageType } = dbus

  try {
    bus = dbus.sessionBus()

    bus.on('error', (err) => {
      console.warn('D-Bus error:', err.message)
    })

    bus.on('disconnect', () => {
      console.warn('D-Bus session bus disconnected')
      cleanup()
      onDisconnect?.()
    })

    // Intercept all incoming messages on the underlying connection.
    // This lets us capture Response signals (on dynamic Request paths)
    // and Activated/Deactivated signals without needing proxy objects
    // for every path.
    signalHandler = (msg) => {
      if (msg.type !== MessageType.SIGNAL) return

      // Portal Request Response — resolve the matching waiter
      if (msg.interface === 'org.freedesktop.portal.Request' && msg.member === 'Response') {
        const waiter = responseWaiters.get(msg.path)
        if (waiter) {
          responseWaiters.delete(msg.path)
          waiter(msg.body)
        }
        return
      }

      // GlobalShortcuts Activated / Deactivated
      if (msg.interface === 'org.freedesktop.portal.GlobalShortcuts') {
        const shortcutId = msg.body?.[1]
        if (shortcutId !== activeShortcutId) return
        if (msg.member === 'Activated') onPress()
        else if (msg.member === 'Deactivated') onRelease()
      }
    }
    bus._connection.on('message', signalHandler)

    // Subscribe to the signals we care about so the D-Bus daemon delivers them
    const dbusObj = await bus.getProxyObject('org.freedesktop.DBus', '/org/freedesktop/DBus')
    const dbusIface = dbusObj.getInterface('org.freedesktop.DBus')
    await dbusIface.AddMatch("type='signal',interface='org.freedesktop.portal.Request',member='Response'")
    await dbusIface.AddMatch("type='signal',interface='org.freedesktop.portal.GlobalShortcuts'")

    senderToken = bus.name.slice(1).replace(/\./g, '_')

    // --- CreateSession ---
    const sessionToken = `chatcoal_ptt`
    const createToken = `chatcoal_create`
    const createRequestPath = `/org/freedesktop/portal/desktop/request/${senderToken}/${createToken}`

    const createResponse = waitForResponse(createRequestPath)

    await bus.call(new Message({
      destination: 'org.freedesktop.portal.Desktop',
      path: '/org/freedesktop/portal/desktop',
      interface: 'org.freedesktop.portal.GlobalShortcuts',
      member: 'CreateSession',
      signature: 'a{sv}',
      body: [{
        'handle_token': new Variant('s', createToken),
        'session_handle_token': new Variant('s', sessionToken),
      }],
    }))

    const createResults = await createResponse
    // Use the session handle from the portal's response (authoritative)
    sessionPath = createResults.session_handle.value

    // --- BindShortcuts ---
    await bindShortcut(webKeyCode)

    console.log('GlobalShortcuts portal PTT initialized, session:', sessionPath)
    return true
  } catch (err) {
    console.warn('GlobalShortcuts portal unavailable:', err.message)
    cleanup()
    return false
  }
}

/**
 * Bind (or rebind) the PTT shortcut to a key.
 */
async function bindShortcut(webKeyCode) {
  const { Variant, Message } = dbus
  const accelerator = CODE_TO_ACCELERATOR[webKeyCode] || ''
  // KDE persists shortcut bindings globally by shortcut ID — once
  // 'push-to-talk' is mapped to a key, BindShortcuts ignores new
  // preferred_trigger values. Including the key in the ID forces KDE
  // to treat each key as a fresh shortcut it hasn't seen before.
  activeShortcutId = `ptt-${webKeyCode}`
  const bindToken = `chatcoal_bind_${++bindCounter}`
  const bindRequestPath = `/org/freedesktop/portal/desktop/request/${senderToken}/${bindToken}`

  const bindResponse = waitForResponse(bindRequestPath)

  await bus.call(new Message({
    destination: 'org.freedesktop.portal.Desktop',
    path: '/org/freedesktop/portal/desktop',
    interface: 'org.freedesktop.portal.GlobalShortcuts',
    member: 'BindShortcuts',
    signature: 'oa(sa{sv})sa{sv}',
    body: [
      sessionPath,
      [[activeShortcutId, {
        'description': new Variant('s', 'Push to Talk'),
        'preferred_trigger': new Variant('s', accelerator),
      }]],
      '',
      { 'handle_token': new Variant('s', bindToken) },
    ],
  }))

  const bindResults = await bindResponse
  const shortcuts = bindResults?.shortcuts?.value
  if (shortcuts?.length) {
    const trigger = shortcuts[0]?.[1]?.trigger_description?.value
    if (trigger) console.log('Portal PTT bound to:', trigger)
  }
}

/**
 * Stop the portal PTT and clean up D-Bus resources.
 */
export function stopPortalPtt() {
  cleanup()
}

function cleanup() {
  responseWaiters.clear()
  if (bus) {
    if (signalHandler && bus._connection) {
      bus._connection.removeListener('message', signalHandler)
    }
    try { bus.disconnect() } catch {}
    bus = null
  }
  signalHandler = null
  sessionPath = null
  senderToken = null
  activeShortcutId = null
}
