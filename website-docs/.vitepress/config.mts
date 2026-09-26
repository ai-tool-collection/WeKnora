import { defineConfig } from 'vitepress'

const repo = 'https://github.com/ai-tool-collection/WeKnora'

export default defineConfig({
  title: 'Knowledge Hub',
  description: 'Self-hosted knowledge base and agent documentation',
  lang: 'en-US',
  base: '/docs/',
  cleanUrls: true,
  lastUpdated: true,
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/01-getting-started/03-quickstart' },
      { text: 'Models', link: '/03-features/06-models' },
      { text: 'API', link: '/04-api/01-api-overview' },
    ],
    sidebar: [
      { text: 'Start', items: [
        { text: 'Overview', link: '/' },
        { text: 'Quickstart', link: '/01-getting-started/03-quickstart' },
        { text: 'Troubleshooting', link: '/01-getting-started/05-troubleshooting' },
      ] },
      { text: 'Architecture', items: [{ text: 'Overview', link: '/02-architecture/01-overview' }] },
      { text: 'Features', items: [
        { text: 'Tenants and access', link: '/03-features/01-tenant-auth' },
        { text: 'Models', link: '/03-features/06-models' },
        { text: 'Knowledge graph', link: '/03-features/09-knowledge-graph' },
        { text: 'Messaging', link: '/03-features/12-im-integration' },
        { text: 'Skills and sandbox', link: '/03-features/22-skills-sandbox' },
      ] },
      { text: 'API', items: [{ text: 'Overview', link: '/04-api/01-api-overview' }] },
      { text: 'Clients', items: [{ text: 'Web frontend', link: '/05-clients/01-frontend' }] },
      { text: 'Development', items: [
        { text: 'Guide', link: '/06-development/01-dev-guide' },
        { text: 'Sandbox deployment', link: '/06-development/04-sandbox-deployment' },
      ] },
    ],
    socialLinks: [{ icon: 'github', link: repo }],
    editLink: { pattern: `${repo}/edit/main/website-docs/:path` },
    search: { provider: 'local' },
    footer: { message: 'Knowledge Hub · MIT License' },
  },
})
