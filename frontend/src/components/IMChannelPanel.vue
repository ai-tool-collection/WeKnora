<template>
  <div class="im-panel">
    <div class="channels-header">
      <span>{{ $t('agentEditor.im.channelsTitle') }}</span>
      <IntegrationsAgentFilter v-model="filterAgentId" :agents="agents" />
      <t-button v-if="authStore.hasRole('admin')" theme="primary" @click="openCreate">
        {{ $t('agentEditor.im.addChannel') }}
      </t-button>
    </div>

    <t-loading :loading="loading">
      <t-empty v-if="!loading && channels.length === 0" :description="$t('agentEditor.im.empty')" />
      <div v-else class="channel-grid">
        <div v-for="channel in channels" :key="channel.id" class="channel-card">
          <img :src="platformLogo(channel.platform)" :alt="channel.platform" class="platform-logo" />
          <div class="channel-info">
            <strong>{{ channel.name || platformLabel(channel.platform) }}</strong>
            <span>{{ agentName(channel.agent_id) }} · {{ platformLabel(channel.platform) }}</span>
          </div>
          <t-tag v-if="!channel.enabled" theme="warning" variant="light">
            {{ $t('agentEditor.im.disabled') }}
          </t-tag>
          <t-button v-if="authStore.hasRole('admin')" variant="text" @click="openEdit(channel)">
            {{ $t('common.edit') }}
          </t-button>
          <t-button v-if="authStore.hasRole('admin')" variant="text" @click="toggle(channel)">
            {{ channel.enabled ? $t('common.off') : $t('common.on') }}
          </t-button>
          <t-popconfirm v-if="authStore.hasRole('admin')" :content="$t('agentEditor.im.deleteConfirm')"
            @confirm="remove(channel.id)">
            <t-button variant="text" theme="danger">{{ $t('common.delete') }}</t-button>
          </t-popconfirm>
        </div>
      </div>
    </t-loading>

    <SettingDrawer v-model:visible="visible" :title="editingId ? $t('common.edit') : $t('agentEditor.im.addChannel')"
      storage-key="setting-drawer:im-channel" width="560px" :confirm-loading="saving"
      :hide-footer="!authStore.hasRole('admin')" @confirm="save">
      <div class="form-item">
        <label>{{ $t('integrations.boundAgent') }}</label>
        <t-select v-model="form.target_agent_id" :options="agentOptions" filterable />
      </div>
      <div class="form-item">
        <label>{{ $t('agentEditor.im.platform') }}</label>
        <t-select v-model="form.platform" :disabled="!!editingId" :options="platformOptions"
          @change="onPlatformChange" />
      </div>
      <div class="form-item">
        <label>{{ $t('agentEditor.im.channelName') }}</label>
        <t-input v-model="form.name" />
      </div>
      <div class="form-item">
        <label>{{ $t('agentEditor.im.mode') }}</label>
        <t-select v-model="form.mode" :options="modeOptions" />
      </div>
      <div class="form-item">
        <label>{{ $t('agentEditor.im.outputMode') }}</label>
        <t-select v-model="form.output_mode" :options="outputOptions" />
      </div>
      <div class="form-item">
        <label>{{ $t('agentEditor.im.replyLanguage') }}</label>
        <t-select v-model="form.locale" :options="localeOptions" />
      </div>
      <div class="form-item">
        <label>{{ $t('agentEditor.im.sessionMode') }}</label>
        <t-select v-model="form.session_mode" :options="sessionOptions" />
      </div>
      <div class="form-item">
        <label>{{ $t('agentEditor.im.knowledgeBase') }}</label>
        <t-select v-model="form.knowledge_base_id" :options="knowledgeBaseOptions" clearable />
      </div>

      <template v-if="form.platform === 'slack'">
        <div class="form-item">
          <label>Bot Token</label>
          <t-input v-model="form.credentials.bot_token" type="password" placeholder="xoxb-..." />
        </div>
        <div v-if="form.mode === 'websocket'" class="form-item">
          <label>App Token</label>
          <t-input v-model="form.credentials.app_token" type="password" placeholder="xapp-..." />
        </div>
        <div v-else class="form-item">
          <label>Signing Secret</label>
          <t-input v-model="form.credentials.signing_secret" type="password" />
        </div>
      </template>
      <template v-else-if="form.platform === 'telegram'">
        <div class="form-item">
          <label>Bot Token</label>
          <t-input v-model="form.credentials.bot_token" type="password" />
        </div>
        <div v-if="form.mode === 'webhook'" class="form-item">
          <label>Secret Token</label>
          <t-input v-model="form.credentials.secret_token" type="password" />
        </div>
      </template>
      <template v-else>
        <div class="form-item">
          <label>Site URL</label>
          <t-input v-model="form.credentials.site_url" placeholder="https://mattermost.example.com" />
        </div>
        <div class="form-item">
          <label>Bot Token</label>
          <t-input v-model="form.credentials.bot_token" type="password" />
        </div>
        <div class="form-item">
          <label>Outgoing Webhook Token</label>
          <t-input v-model="form.credentials.outgoing_token" type="password" />
        </div>
        <div class="form-item">
          <label>Bot User ID</label>
          <t-input v-model="form.credentials.bot_user_id" />
        </div>
        <div class="form-item">
          <label>{{ $t('agentEditor.im.mattermostPostToMain') }}</label>
          <t-switch v-model="form.credentials.post_to_main" />
        </div>
      </template>
      <div v-if="editingId && form.mode === 'webhook'" class="form-item">
        <label>Webhook URL</label>
        <t-input :value="callbackURL" readonly />
        <t-button variant="text" @click="copyCallback">{{ $t('common.copy') }}</t-button>
      </div>
      <div v-if="editingId" class="form-item">
        <label>{{ $t('agentEditor.im.enabled') }}</label>
        <t-switch v-model="form.enabled" />
      </div>
    </SettingDrawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { MessagePlugin } from 'tdesign-vue-next';
import SettingDrawer from '@/components/settings/SettingDrawer.vue';
import IntegrationsAgentFilter from '@/components/IntegrationsAgentFilter.vue';
import { copyWithToast } from '@/utils/clipboard';
import { normalizeOptionalString } from '@/utils/optionalString';
import { useAuthStore } from '@/stores/auth';
import { useChatResourcesStore } from '@/stores/chatResources';
import {
  listIMChannels, listAllIMChannels, createIMChannel, updateIMChannel,
  toggleIMChannel, deleteIMChannel,
  type IMChannel, type IMChannelOverview, type CustomAgent,
} from '@/api/agent';
import slackLogo from '@/assets/img/im/slack.svg';
import telegramLogo from '@/assets/img/im/telegram.svg';
import mattermostLogo from '@/assets/img/im/mattermost.svg';

type Platform = IMChannel['platform'];
type ChannelForm = {
  target_agent_id: string;
  platform: Platform;
  name: string;
  enabled: boolean;
  mode: IMChannel['mode'];
  output_mode: IMChannel['output_mode'];
  locale: NonNullable<IMChannel['locale']>;
  session_mode: NonNullable<IMChannel['session_mode']>;
  knowledge_base_id: string;
  credentials: Record<string, any>;
};

const { t } = useI18n();
const authStore = useAuthStore();
const filterAgentId = defineModel<string>('filterAgentId', { default: '' });
const agents = ref<CustomAgent[]>([]);
const allChannels = ref<IMChannelOverview[]>([]);
const knowledgeBases = ref<{ id: string; name: string }[]>([]);
const loading = ref(false);
const saving = ref(false);
const visible = ref(false);
const editingId = ref('');
const logos: Record<Platform, string> = {
  slack: slackLogo, telegram: telegramLogo, mattermost: mattermostLogo,
};
const platformLogo = (platform: string) => logos[platform as Platform] || '';
const platformLabel = (platform: string) => t('agentEditor.im.' + platform);
const channels = computed(() => allChannels.value.filter((channel) =>
  platformLogo(channel.platform) && (!filterAgentId.value || channel.agent_id === filterAgentId.value),
));
const agentOptions = computed(() => agents.value.map((agent) => ({ value: agent.id, label: agent.name })));
const agentName = (id: string) => agents.value.find((agent) => agent.id === id)?.name || '';
const platformOptions = computed(() => ([
  { value: 'slack', label: platformLabel('slack') },
  { value: 'telegram', label: platformLabel('telegram') },
  { value: 'mattermost', label: platformLabel('mattermost') },
]));
const modeOptions = computed(() => form.value.platform === 'mattermost'
  ? [{ value: 'webhook', label: 'Webhook' }]
  : form.value.platform === 'telegram'
    ? [{ value: 'webhook', label: 'Webhook' }, { value: 'websocket', label: 'Long polling' }]
    : [{ value: 'websocket', label: 'Socket Mode' }, { value: 'webhook', label: 'Webhook' }]);
const outputOptions = [
  { value: 'stream', label: 'Stream' }, { value: 'full', label: 'Full' },
];
const localeOptions = [
  { value: '', label: 'Default' }, { value: 'en-US', label: 'English' },
  { value: 'ja-JP', label: '日本語' }, { value: 'ko-KR', label: '한국어' },
  { value: 'ru-RU', label: 'Русский' },
];
const sessionOptions = [
  { value: 'user', label: 'User' }, { value: 'thread', label: 'Thread' },
];
const knowledgeBaseOptions = computed(() => [
  { value: '', label: 'None' },
  ...knowledgeBases.value.map((kb) => ({ value: kb.id, label: kb.name })),
]);
function emptyForm(): ChannelForm {
  return {
    target_agent_id: filterAgentId.value || '',
    platform: 'slack',
    name: '',
    enabled: true,
    mode: 'websocket',
    output_mode: 'stream',
    locale: '',
    session_mode: 'user',
    knowledge_base_id: '',
    credentials: {},
  };
}
const form = ref<ChannelForm>(emptyForm());
const callbackURL = computed(() => editingId.value
  ? window.location.origin + '/api/v1/im/callback/' + editingId.value : '');

async function load() {
  loading.value = true;
  try {
    const resources = useChatResourcesStore();
    const [channelResult] = await Promise.all([
      listAllIMChannels(), resources.ensureAgents(), resources.ensureKnowledgeBases(),
    ]);
    allChannels.value = channelResult.data || [];
    agents.value = resources.agents as CustomAgent[];
    knowledgeBases.value = resources.rawKnowledgeBases.map((kb: any) => ({ id: kb.id, name: kb.name }));
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('common.operationFailed'));
  } finally {
    loading.value = false;
  }
}
function onPlatformChange() {
  form.value.mode = form.value.platform === 'mattermost' ? 'webhook' : 'websocket';
  if (form.value.platform === 'telegram') form.value.mode = 'websocket';
  form.value.credentials = {};
}
function openCreate() {
  editingId.value = '';
  form.value = emptyForm();
  visible.value = true;
}
async function openEdit(channel: IMChannelOverview) {
  try {
    const result = await listIMChannels(channel.agent_id);
    const full = result.data.find((item) => item.id === channel.id);
    if (!full) throw new Error(t('common.operationFailed'));
    editingId.value = full.id;
    form.value = {
      target_agent_id: full.agent_id, platform: full.platform, name: full.name,
      enabled: full.enabled, mode: full.mode, output_mode: full.output_mode,
      locale: full.locale || '', session_mode: full.session_mode || 'user',
      knowledge_base_id: full.knowledge_base_id || '', credentials: { ...full.credentials },
    };
    visible.value = true;
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('common.operationFailed'));
  }
}
async function save() {
  if (!form.value.target_agent_id) {
    MessagePlugin.warning(t('integrations.selectAgentHint'));
    return;
  }
  saving.value = true;
  try {
    const payload = {
      name: form.value.name.trim() || platformLabel(form.value.platform),
      mode: form.value.mode, output_mode: form.value.output_mode,
      locale: form.value.locale, session_mode: form.value.session_mode,
      knowledge_base_id: normalizeOptionalString(form.value.knowledge_base_id),
      credentials: form.value.credentials, enabled: form.value.enabled,
    };
    if (editingId.value) {
      await updateIMChannel(editingId.value, { ...payload, agent_id: form.value.target_agent_id });
    } else {
      await createIMChannel(form.value.target_agent_id, { ...payload, platform: form.value.platform });
    }
    visible.value = false;
    MessagePlugin.success(t(editingId.value ? 'common.updateSuccess' : 'common.createSuccess'));
    await load();
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('common.operationFailed'));
  } finally {
    saving.value = false;
  }
}
async function toggle(channel: IMChannelOverview) {
  try {
    await toggleIMChannel(channel.id);
    await load();
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('common.operationFailed'));
  }
}
async function remove(id: string) {
  try {
    await deleteIMChannel(id);
    MessagePlugin.success(t('common.deleteSuccess'));
    await load();
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('common.operationFailed'));
  }
}
async function copyCallback() {
  await copyWithToast(callbackURL.value, 'common.copySuccess');
}
onMounted(load);
</script>

<style scoped lang="less">
.im-panel { display: flex; flex-direction: column; gap: 18px; }
.channels-header { display: flex; align-items: center; gap: 12px; font-weight: 600; }
.channels-header > button { margin-left: auto; }
.channel-grid { display: grid; gap: 10px; }
.channel-card { display: flex; align-items: center; gap: 12px; padding: 14px; border: 1px solid var(--td-component-border); border-radius: var(--app-radius-md); }
.platform-logo { width: 28px; height: 28px; }
.channel-info { display: flex; flex: 1; flex-direction: column; gap: 3px; }
.channel-info span { color: var(--td-text-color-secondary); font-size: 12px; }
.form-item { display: flex; flex-direction: column; gap: 6px; margin: 0 0 16px; }
.form-item label { font-weight: 500; }
</style>
