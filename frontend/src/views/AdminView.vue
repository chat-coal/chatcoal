<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import api from '@/services/api.service'

const router = useRouter()
const authStore = useAuthStore()
const toastStore = useToastStore()

const loading = ref(true)
const federationAvailable = ref(false)
const defaultPolicy = ref('open')
const instances = ref([])
const saving = ref(false)

// Add form
const newDomain = ref('')
const newPolicy = ref('allow')
const newNote = ref('')
const adding = ref(false)

// Confirm remove
const removingDomain = ref(null)

// Reports
const reports = ref([])
const reportsLoading = ref(false)
const reportsFilter = ref('')
const reportsHasMore = ref(true)

onMounted(async () => {
  if (!authStore.dbUser?.is_site_admin) {
    router.replace('/channels/@me')
    return
  }
  try {
    const data = await api.getFederationPolicy()
    defaultPolicy.value = data.default_policy
    instances.value = data.instances || []
    federationAvailable.value = true
  } catch {
    // Federation not configured — skip section silently
  } finally {
    loading.value = false
  }
  loadReports()
})

async function loadReports(append = false) {
  reportsLoading.value = true
  try {
    const before = append && reports.value.length ? reports.value[reports.value.length - 1].id : null
    const data = await api.getReports(reportsFilter.value, before)
    if (append) {
      reports.value = [...reports.value, ...data]
    } else {
      reports.value = data
    }
    reportsHasMore.value = data.length >= 50
  } catch {
    toastStore.add('Failed to load reports')
  } finally {
    reportsLoading.value = false
  }
}

function changeReportsFilter(status) {
  reportsFilter.value = status
  loadReports()
}

async function setReportStatus(reportId, status) {
  try {
    const updated = await api.updateReportStatus(reportId, status)
    const idx = reports.value.findIndex(r => r.id === reportId)
    if (idx >= 0) reports.value[idx] = { ...reports.value[idx], status: updated.status }
  } catch {
    toastStore.add('Failed to update report')
  }
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const targetTypeLabels = { user: 'User', message: 'Message', dm_message: 'DM Message', forum_post: 'Forum Post' }

async function saveDefaultPolicy() {
  saving.value = true
  try {
    await api.updateFederationPolicy(defaultPolicy.value)
    toastStore.add('Default policy updated')
  } catch {
    toastStore.add('Failed to update policy')
  } finally {
    saving.value = false
  }
}

async function addInstance() {
  if (!newDomain.value.trim()) return
  adding.value = true
  try {
    const p = await api.addInstancePolicy(newDomain.value.trim(), newPolicy.value, newNote.value.trim())
    // Upsert into local list
    const idx = instances.value.findIndex(i => i.domain === p.domain)
    if (idx >= 0) instances.value[idx] = p
    else instances.value.unshift(p)
    newDomain.value = ''
    newNote.value = ''
    newPolicy.value = 'allow'
  } catch (e) {
    toastStore.add(e.response?.data?.error || 'Failed to add policy')
  } finally {
    adding.value = false
  }
}

async function removeInstance(domain) {
  try {
    await api.removeInstancePolicy(domain)
    instances.value = instances.value.filter(i => i.domain !== domain)
    removingDomain.value = null
  } catch {
    toastStore.add('Failed to remove policy')
  }
}

function goBack() {
  router.push('/channels/@me')
}
</script>

<template>
  <div class="min-h-screen bg-[var(--bg)] text-[var(--text-1)]">
    <!-- Header -->
    <div class="max-w-3xl mx-auto px-6 pt-8 pb-4">
      <button @click="goBack" class="flex items-center gap-1.5 text-[var(--text-3)] hover:text-[var(--text-1)] text-sm mb-6 transition-colors duration-150 cursor-pointer">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
        </svg>
        Back to app
      </button>
      <h1 class="text-2xl font-bold mb-1">Admin</h1>
      <p class="text-[var(--text-3)] text-sm">Content moderation{{ federationAvailable ? ' and federation policy' : '' }}.</p>
    </div>

    <div v-if="loading" class="max-w-3xl mx-auto px-6 py-12 text-center">
      <div class="inline-block w-6 h-6 border-2 border-[var(--text-4)] border-t-[#E8521A] rounded-full animate-spin"></div>
    </div>

    <div v-else class="max-w-3xl mx-auto px-6 pb-12 space-y-8">
      <!-- Default Policy (federation only) -->
      <section v-if="federationAvailable" class="bg-[var(--surface)] border border-[var(--surface-border)] rounded-2xl p-6">
        <h2 class="text-lg font-semibold mb-1">Default Policy</h2>
        <p class="text-[var(--text-3)] text-sm mb-4">Applies to instances without an explicit rule below.</p>
        <div class="flex items-center gap-3">
          <button
            @click="defaultPolicy = 'open'"
            class="px-4 py-2 rounded-xl text-sm font-semibold transition-all duration-150 cursor-pointer"
            :class="defaultPolicy === 'open'
              ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
              : 'bg-[var(--surface-2)] text-[var(--text-3)] border border-[var(--surface-border)] hover:border-[var(--text-4)]'"
          >Open</button>
          <button
            @click="defaultPolicy = 'closed'"
            class="px-4 py-2 rounded-xl text-sm font-semibold transition-all duration-150 cursor-pointer"
            :class="defaultPolicy === 'closed'
              ? 'bg-red-500/15 text-red-400 border border-red-500/30'
              : 'bg-[var(--surface-2)] text-[var(--text-3)] border border-[var(--surface-border)] hover:border-[var(--text-4)]'"
          >Closed</button>
          <button
            @click="saveDefaultPolicy"
            :disabled="saving"
            class="ml-auto px-5 py-2 rounded-xl text-sm font-semibold text-white bg-[#E8521A] hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/15 transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >{{ saving ? 'Saving...' : 'Save' }}</button>
        </div>
      </section>

      <!-- Instance Rules (federation only) -->
      <section v-if="federationAvailable" class="bg-[var(--surface)] border border-[var(--surface-border)] rounded-2xl p-6">
        <h2 class="text-lg font-semibold mb-1">Instance Rules</h2>
        <p class="text-[var(--text-3)] text-sm mb-5">Explicit allow/block overrides the default policy.</p>

        <!-- Add form -->
        <div class="flex flex-wrap gap-2 mb-5">
          <input
            v-model="newDomain"
            placeholder="instance.example.com"
            @keyup.enter="addInstance"
            class="flex-1 min-w-[180px] bg-[var(--surface-3)] text-[var(--text-1)] text-sm px-3 py-2.5 rounded-xl placeholder-[var(--text-4)] border border-[var(--surface-border)] focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
          />
          <div class="relative">
            <select
              v-model="newPolicy"
              class="appearance-none bg-[var(--surface-3)] text-[var(--text-1)] text-sm px-3 py-2.5 pr-9 rounded-xl border border-[var(--surface-border)] focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150 cursor-pointer"
            >
              <option value="allow">Allow</option>
              <option value="block">Block</option>
            </select>
            <svg class="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-3)]" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd"/></svg>
          </div>
          <input
            v-model="newNote"
            placeholder="Note (optional)"
            @keyup.enter="addInstance"
            class="flex-1 min-w-[120px] bg-[var(--surface-3)] text-[var(--text-1)] text-sm px-3 py-2.5 rounded-xl placeholder-[var(--text-4)] border border-[var(--surface-border)] focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
          />
          <button
            @click="addInstance"
            :disabled="adding || !newDomain.trim()"
            class="px-5 py-2.5 rounded-xl text-sm font-semibold text-white bg-[#E8521A] hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/15 transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >{{ adding ? 'Adding...' : 'Add' }}</button>
        </div>

        <!-- Instance list -->
        <div v-if="instances.length === 0" class="text-[var(--text-4)] text-sm text-center py-6">
          No instance-specific rules configured.
        </div>
        <div v-else class="divide-y divide-[var(--surface-border)]">
          <div
            v-for="inst in instances"
            :key="inst.domain"
            class="flex items-center gap-3 py-3"
          >
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium truncate">{{ inst.domain }}</p>
              <p v-if="inst.note" class="text-[var(--text-4)] text-xs truncate mt-0.5">{{ inst.note }}</p>
            </div>
            <span
              class="shrink-0 px-2.5 py-1 rounded-lg text-xs font-semibold"
              :class="inst.policy === 'allow'
                ? 'bg-emerald-500/15 text-emerald-400'
                : 'bg-red-500/15 text-red-400'"
            >{{ inst.policy }}</span>
            <button
              v-if="removingDomain !== inst.domain"
              @click="removingDomain = inst.domain"
              class="shrink-0 text-[var(--text-4)] hover:text-red-400 p-1 rounded transition-colors duration-150 cursor-pointer"
              title="Remove"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
              </svg>
            </button>
            <div v-else class="flex items-center gap-1.5 shrink-0">
              <button
                @click="removeInstance(inst.domain)"
                class="text-xs font-semibold text-red-400 hover:text-red-300 px-2 py-1 rounded-lg bg-red-500/10 hover:bg-red-500/20 transition-all duration-150 cursor-pointer"
              >Remove</button>
              <button
                @click="removingDomain = null"
                class="text-xs text-[var(--text-4)] hover:text-[var(--text-2)] px-2 py-1 cursor-pointer transition-colors duration-150"
              >Cancel</button>
            </div>
          </div>
        </div>
      </section>

      <!-- Reports -->
      <section class="bg-[var(--surface)] border border-[var(--surface-border)] rounded-2xl p-6">
        <h2 class="text-lg font-semibold mb-1">Reports</h2>
        <p class="text-[var(--text-3)] text-sm mb-5">User-submitted reports for review.</p>

        <!-- Filter tabs -->
        <div class="flex items-center gap-2 mb-5">
          <button
            v-for="f in [{ label: 'All', value: '' }, { label: 'Pending', value: 'pending' }, { label: 'Reviewed', value: 'reviewed' }, { label: 'Dismissed', value: 'dismissed' }]"
            :key="f.value"
            @click="changeReportsFilter(f.value)"
            class="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all duration-150 cursor-pointer"
            :class="reportsFilter === f.value
              ? 'bg-[#E8521A]/15 text-[#E8521A] border border-[#E8521A]/30'
              : 'bg-[var(--surface-2)] text-[var(--text-3)] border border-[var(--surface-border)] hover:border-[var(--text-4)]'"
          >{{ f.label }}</button>
        </div>

        <!-- Reports list -->
        <div v-if="reportsLoading && !reports.length" class="text-center py-8">
          <div class="inline-block w-5 h-5 border-2 border-[var(--text-4)] border-t-[#E8521A] rounded-full animate-spin"></div>
        </div>
        <div v-else-if="reports.length === 0" class="text-[var(--text-4)] text-sm text-center py-6">
          No reports found.
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="report in reports"
            :key="report.id"
            class="bg-[var(--surface-2)] border border-[var(--surface-border)] rounded-xl p-4"
          >
            <div class="flex items-start gap-3">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1.5 flex-wrap">
                  <span
                    class="px-2 py-0.5 rounded-md text-[10px] font-bold uppercase"
                    :class="{
                      'bg-amber-500/15 text-amber-400': report.status === 'pending',
                      'bg-emerald-500/15 text-emerald-400': report.status === 'reviewed',
                      'bg-[var(--surface-3)] text-[var(--text-4)]': report.status === 'dismissed',
                    }"
                  >{{ report.status }}</span>
                  <span class="px-2 py-0.5 rounded-md text-[10px] font-semibold bg-[var(--surface-3)] text-[var(--text-3)]">
                    {{ targetTypeLabels[report.target_type] || report.target_type }}
                  </span>
                  <span class="text-[var(--text-4)] text-[11px]">{{ formatDate(report.created_at) }}</span>
                </div>
                <p class="text-sm text-[var(--text-2)] mb-1">{{ report.reason }}</p>
                <div class="flex items-center gap-2 text-[11px] text-[var(--text-4)]">
                  <span>Reported by: <span class="text-[var(--text-3)] font-medium">{{ report.reporter?.display_name || 'Unknown' }}</span></span>
                  <span>&middot;</span>
                  <span>Target ID: <span class="font-mono text-[var(--text-3)]">{{ report.target_id }}</span></span>
                </div>
              </div>

              <!-- Actions -->
              <div v-if="report.status === 'pending'" class="flex items-center gap-1.5 shrink-0">
                <button
                  @click="setReportStatus(report.id, 'reviewed')"
                  class="px-2.5 py-1.5 rounded-lg text-xs font-semibold text-emerald-400 bg-emerald-500/10 hover:bg-emerald-500/20 transition-all duration-150 cursor-pointer"
                  title="Mark as reviewed"
                >Review</button>
                <button
                  @click="setReportStatus(report.id, 'dismissed')"
                  class="px-2.5 py-1.5 rounded-lg text-xs font-semibold text-[var(--text-4)] bg-[var(--surface-3)] hover:bg-[var(--surface-border)] transition-all duration-150 cursor-pointer"
                  title="Dismiss"
                >Dismiss</button>
              </div>
              <div v-else class="shrink-0">
                <button
                  @click="setReportStatus(report.id, 'pending')"
                  class="px-2.5 py-1.5 rounded-lg text-xs text-[var(--text-4)] hover:text-[var(--text-3)] bg-[var(--surface-3)] hover:bg-[var(--surface-border)] transition-all duration-150 cursor-pointer"
                >Reopen</button>
              </div>
            </div>
          </div>

          <!-- Load more -->
          <div v-if="reportsHasMore" class="text-center pt-2">
            <button
              @click="loadReports(true)"
              :disabled="reportsLoading"
              class="text-[#E8521A] hover:text-[#D44818] text-sm font-semibold cursor-pointer transition-colors duration-150 disabled:opacity-50"
            >{{ reportsLoading ? 'Loading...' : 'Load more' }}</button>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
