import DefaultTheme from 'vitepress/theme'
import './style.css'
import StickyBanner from './StickyBanner.vue'
import { h } from 'vue'

export default {
  extends: DefaultTheme,
  Layout: () => {
    return h(DefaultTheme.Layout, null, {
      'layout-top': () => h(StickyBanner)
    })
  }
}
