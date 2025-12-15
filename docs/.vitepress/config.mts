import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Mushak",
  description: "Mushak (Persian for Rocket) is a zero-config deployment tool for your own server.",
  head: [['link', { rel: 'icon', href: '/logo-mushak.svg' }]],
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: '/logo-mushak.svg',
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Guide', link: '/guide/what-is-mushak' },
      { text: 'Reference', link: '/guide/commands' }
    ],

    sidebar: [
      {
        text: 'Introduction',
        items: [
          { text: 'What is Mushak?', link: '/guide/what-is-mushak' },
          { text: 'Getting Started', link: '/guide/getting-started' },
        ]
      },
      {
        text: 'Usage',
        items: [
          { text: 'Deployment', link: '/guide/deployment' },
          { text: 'Configuration', link: '/guide/configuration' },
          { text: 'Troubleshooting', link: '/guide/troubleshooting' },
        ]
      },
      {
        text: 'Internals',
        items: [
          { text: 'Architecture & Details', link: '/guide/architecture' },
        ]
      },
      {
        text: 'Reference',
        items: [
          { text: 'CLI Commands', link: '/guide/commands' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/hmontazeri/mushak' }
    ]
  }
})
