import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const source = readFileSync(
  new URL('./SandboxConfigEditorDrawer.vue', import.meta.url), 'utf8')

test('network policy lives in the runtime step, right below the runtime config', () => {
  const runtimeSections = source.indexOf("currentStepKey === 'runtime'")
  const networkSection = source.indexOf('settings.sandbox.sectionNetwork')
  const envSection = source.indexOf('settings.sandbox.sectionEnvironment')

  assert.ok(runtimeSections !== -1, 'runtime step must exist')
  assert.ok(networkSection !== -1, 'network policy section must exist')
  assert.ok(
    networkSection > runtimeSections && networkSection < envSection,
    'network policy must sit between the runtime config and the env vars',
  )
  assert.ok(
    !source.includes("currentStepKey === 'network'"),
    'network policy must not become a fourth wizard step',
  )
})

test('runtime numbers are laid out two per row with per-field tips', () => {
  const runtimeBlock = source.slice(
    source.indexOf('settings.sandbox.sectionRuntime'),
    source.indexOf('settings.sandbox.sectionNetwork'),
  )
  assert.ok(
    runtimeBlock.includes('form-grid form-grid--two'),
    'timeouts must be compressed into the two-column grid',
  )
  // Each number keeps its own sentence; the previous layout put those in
  // block paragraphs, which is what made the step so tall.
  assert.ok(
    runtimeBlock.includes(":tips=\"$t('settings.sandbox.httpTimeoutHelp')\""),
    'http timeout keeps its own explanation, now inline',
  )
  assert.ok(
    runtimeBlock.includes(":tips=\"$t('settings.sandbox.sandboxTtlHelp')\""),
    'sandbox TTL keeps its own explanation, now inline',
  )
  for (const key of [
    'dockerCpuLimitHelp',
    'dockerMemoryLimitHelp',
    'dockerPidsLimitHelp',
  ]) {
    assert.ok(
      runtimeBlock.includes(`:tips="$t('settings.sandbox.${key}')"`),
      `${key} must remain attached to its compressed field`,
    )
  }
})

test('docker network mode moved into the network policy section', () => {
  const networkBlock = source.slice(source.indexOf('settings.sandbox.sectionNetwork'))
  assert.ok(
    networkBlock.includes('settings.sandbox.dockerNetworkMode'),
    'the docker bridge/none selector belongs with the other network controls',
  )
  const runtimeBlock = source.slice(
    source.indexOf('settings.sandbox.sectionRuntime'),
    source.indexOf('settings.sandbox.sectionNetwork'),
  )
  assert.ok(
    !runtimeBlock.includes('settings.sandbox.dockerNetworkMode'),
    'and must no longer be duplicated in the runtime config',
  )
})

test('inbound is always credential-required and never shown', () => {
  assert.doesNotMatch(
    source,
    /settings\.sandbox\.inboundAccess/,
    'the inbound radio must not appear in the form',
  )
  assert.doesNotMatch(
    source,
    /allowPublicInbound/,
    'the form must not keep a public-inbound control',
  )
  const payloadBlock = source.slice(
    source.indexOf('function collectNetworkPolicy'),
    source.indexOf('function close'),
  )
  assert.doesNotMatch(
    payloadBlock,
    /allow_public_inbound/,
    'saves must omit allow_public_inbound so the zero value (require credentials) is stored',
  )
  assert.match(
    source,
    /const denyEgressByDefault = ref\(false\)/,
    'egress allowed is the default',
  )
})




test('docker payload omits hidden allow and deny lists', () => {
  const payloadBlock = source.slice(
    source.indexOf('function collectNetworkPolicy'),
    source.indexOf('function close'),
  )
  assert.match(
    payloadBlock,
    /if \(backend\.value === 'docker'\) \{\s*return policy/,
    'Docker must not persist remote backend radios or allow/deny rows',
  )
  assert.match(
    payloadBlock,
    /if \(backend\.value !== 'docker'\) \{[^}]*policy\.allow_out = allowOut[^}]*policy\.deny_out = denyOut[^}]*\}/,
    'Docker must not send allow/deny rows that its form hides',
  )
})

test('docker hides egress radios that it cannot honour', () => {
  const networkBlock = source.slice(
    source.indexOf('settings.sandbox.sectionNetwork'),
    source.indexOf('settings.sandbox.sectionEnvironment'),
  )
  assert.match(
    networkBlock,
    /v-if="backend !== 'docker'"[\s\S]*settings\.sandbox\.egressDefault/,
    'egress radios must not render for Docker',
  )
})



test('policy-restricted egress does not claim full outbound verification', () => {
  assert.match(
    source,
    /EGRESS_RESTRICTED_REASON = 'egress_restricted_by_policy'/,
    'the hint must recognize the skip reason the check API returns',
  )
  assert.match(
    source,
    /settings\.sandbox\.checkScopePolicyRestricted/,
    'deny-by-default must not reuse the "outbound was verified" copy',
  )
  const hintBlock = source.slice(
    source.indexOf('const checkScopeHint = computed'),
    source.indexOf('function checkDetail'),
  )
  const restrictedAt = hintBlock.indexOf('checkScopePolicyRestricted')
  const fullAt = hintBlock.indexOf('checkScopeFull')
  assert.ok(
    restrictedAt !== -1 && fullAt !== -1 && restrictedAt < fullAt,
    'policy-restricted copy must win over the full-verification claim',
  )
})
