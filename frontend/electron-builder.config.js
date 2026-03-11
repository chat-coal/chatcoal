/**
 * @type {import('electron-builder').Configuration}
 */
export default {
  appId: 'com.chatcoal.app',
  productName: 'chatcoal',
  directories: {
    output: 'release',
  },
  files: [
    'dist/**/*',
    'electron/**/*',
  ],
  asarUnpack: [
    '**/node_modules/uiohook-napi/**',
  ],
  publish: {
    provider: 'github',
    owner: 'chat-coal',
    repo: 'chatcoal',
    releaseType: 'release',
  },
  mac: {
    target: ['dmg', 'zip'],
    category: 'public.app-category.social-networking',
    icon: 'build/icon.icns',
    hardenedRuntime: true,
    gatekeeperAssess: false,
    entitlements: 'build/entitlements.mac.plist',
    entitlementsInherit: 'build/entitlements.mac.plist',
    notarize: true,
  },
  win: {
    target: ['nsis'],
    icon: 'build/icon.icns',
  },
  linux: {
    target: ['AppImage'],
    category: 'Network;Chat',
    icon: 'build/icon.icns',
  },
}
