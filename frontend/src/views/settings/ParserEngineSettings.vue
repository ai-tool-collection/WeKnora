<template>
  <div class="parser-settings">
    <h2>{{ $t('settings.parser.title') }}</h2>
    <p>{{ $t('settings.parser.description') }}</p>
    <t-loading v-if="loading" size="small" />
    <t-alert v-else-if="error" theme="error" :message="error" />
    <template v-else>
      <div class="engine-list">
        <div v-for="engine in engines" :key="engine.Name" class="engine-row">
          <div><strong>{{ engineName(engine.Name) }}</strong><p>{{ engine.Description }}</p></div>
          <span :class="engine.Available ? 'available' : 'unavailable'">
            {{ engine.Available ? $t('settings.parser.available') : $t('settings.parser.unavailable') }}
          </span>
        </div>
      </div>
      <section v-if="authStore.hasRole('admin')" class="odl-settings">
        <h3>OpenDataLoader PDF</h3>
        <label for="odl-hybrid">Hybrid mode</label>
        <t-select id="odl-hybrid" v-model="config.odl_hybrid">
          <t-option value="off" label="Off" /><t-option value="docling-fast" label="Docling Fast" />
          <t-option value="hancom-ai" label="Hancom AI" />
        </t-select>
        <template v-if="config.odl_hybrid && config.odl_hybrid !== 'off'">
          <label for="odl-url">Hybrid service URL</label>
          <t-input id="odl-url" v-model="config.odl_hybrid_url" placeholder="https://parser.example.com" />
          <label for="odl-mode">Processing mode</label>
          <t-select id="odl-mode" v-model="config.odl_hybrid_mode">
            <t-option value="auto" label="Auto" /><t-option value="full" label="Full" />
          </t-select>
          <label><t-switch v-model="config.odl_hybrid_fallback" /> Fall back to local parsing</label>
        </template>
        <label><t-switch v-model="config.odl_markdown_with_html" /> Keep HTML in Markdown</label>
        <div class="actions">
          <t-button theme="default" :loading="checking" @click="checkConnection">{{ $t('settings.parser.testConnection') }}</t-button>
          <t-button theme="primary" :loading="saving" @click="save">Save</t-button>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { MessagePlugin } from 'tdesign-vue-next'
import { checkParserEngines, getParserEngineConfig, getParserEngines, updateParserEngineConfig, type ParserEngineConfig, type ParserEngineInfo } from '@/api/system'

const { t } = useI18n()
const authStore = useAuthStore()
const engines = ref<ParserEngineInfo[]>([])
const config = ref<ParserEngineConfig>({ odl_hybrid: 'off', odl_hybrid_mode: 'auto' })
const loading = ref(true)
const saving = ref(false)
const checking = ref(false)
const error = ref('')

function engineName(name: string): string {
  const key = `kbSettings.parser.engines.${name}.name`
  const value = t(key)
  return value === key ? name : value
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [catalog, saved] = await Promise.all([getParserEngines(), getParserEngineConfig()])
    engines.value = catalog.data.filter(engine => ['builtin', 'simple', 'anydoc', 'markitdown'].includes(engine.Name))
    config.value = {
      odl_hybrid: saved.data?.odl_hybrid || 'off',
      odl_hybrid_url: saved.data?.odl_hybrid_url || '',
      odl_hybrid_mode: saved.data?.odl_hybrid_mode || 'auto',
      odl_hybrid_fallback: saved.data?.odl_hybrid_fallback ?? false,
      odl_markdown_with_html: saved.data?.odl_markdown_with_html ?? false,
    }
  } catch (cause: any) {
    error.value = cause?.message || t('settings.parser.loadFailed')
  } finally {
    loading.value = false
  }
}

async function checkConnection() {
  checking.value = true
  try {
    const result = await checkParserEngines(config.value)
    engines.value = result.data.filter(engine => ['builtin', 'simple', 'anydoc', 'markitdown'].includes(engine.Name))
    MessagePlugin.success(t('settings.parser.checkSuccess'))
  } catch (cause: any) {
    MessagePlugin.error(cause?.message || t('settings.parser.checkFailed'))
  } finally {
    checking.value = false
  }
}

async function save() {
  saving.value = true
  try {
    await updateParserEngineConfig(config.value)
    MessagePlugin.success(t('settings.parser.saveSuccess'))
    await load()
  } catch (cause: any) {
    MessagePlugin.error(cause?.message || t('settings.parser.saveFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.parser-settings { display: grid; gap: 16px; }
.parser-settings h2, .parser-settings h3, .parser-settings p { margin: 0; }
.engine-list { display: grid; gap: 8px; }
.engine-row { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 14px; border: 1px solid var(--td-border-level-1-color); border-radius: var(--app-radius-md); }
.engine-row p { margin-top: 4px; color: var(--td-text-color-secondary); }
.available { color: var(--td-success-color); }
.unavailable { color: var(--td-warning-color); }
.odl-settings { display: grid; gap: 12px; max-width: 520px; }
.actions { display: flex; gap: 8px; }
</style>
