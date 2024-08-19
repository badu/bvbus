<template>
  <Portal>
    <div v-if="containerVisible" :ref="maskRef" :class="cx('mask')"
         :style="sx('mask', true, { position })" v-bind="ptm('mask')">
      <transition name="p-drawer" @enter="onEnter" @before-leave="onBeforeLeave"
                  @leave="onLeave" @after-leave="onAfterLeave" appear v-bind="ptm('transition')">
        <div v-if="visible" :ref="containerRef" v-focustrap :class="cx('root')" role="complementary" :aria-modal="modal"
             v-bind="ptmi('root')">
          <slot v-if="$slots.container" name="container" :closeCallback="hide"></slot>
          <template v-else>
            <div :ref="headerContainerRef" :class="cx('header')" v-bind="ptm('header')">
              <Button
                  v-if="showCloseIcon"
                  :ref="closeButtonRef"
                  type="button"
                  :class="cx('pcCloseButton')"
                  :aria-label="closeAriaLabel"
                  @click="hide"
                  v-bind="closeButtonProps"
                  :pt="ptm('pcCloseButton')"
                  style="margin-right: 5px;margin-left: 5px;"
                  data-pc-group-section="iconcontainer">
                <template #icon="slotProps">
                  <slot name="closeicon">
                    <component :is="closeIcon ? 'span' : 'TimesIcon'" :class="[closeIcon, slotProps.class]"
                               v-bind="ptm('pcCloseButton')['icon']"></component>
                  </slot>
                </template>
              </Button>
              <slot name="header" :class="cx('title')">
                <div v-if="header" :class="cx('title')" v-bind="ptm('title')">{{ header }}</div>
              </slot>
            </div>
            <div v-if="hasContent" :ref="contentRef" :class="cx('content')" v-bind="ptm('content')">
              <slot></slot>
            </div>
          </template>
        </div>
      </transition>
    </div>
  </Portal>
</template>

<script>
import {addClass, focus} from '@primeuix/utils/dom'
import {ZIndex} from '@primeuix/utils/zindex'
import TimesIcon from '@primevue/icons/times'
import Button from 'primevue/button'
import FocusTrap from 'primevue/focustrap'
import Portal from 'primevue/portal'
import BaseDrawer from './BaseDrawer.vue'

export default {
  name: 'Drawer',
  extends: BaseDrawer,
  inheritAttrs: false,
  emits: ['update:visible', 'show', 'hide', 'after-hide'],
  data() {
    return {
      containerVisible: this.visible
    }
  },
  container: null,
  mask: null,
  content: null,
  headerContainer: null,
  closeButton: null,
  outsideClickListener: null,
  documentKeydownListener: null,
  updated() {
    if (this.visible) {
      this.containerVisible = this.visible
    }
  },
  beforeUnmount() {
    if (this.mask && this.autoZIndex) {
      ZIndex.clear(this.mask)
    }

    this.container = null
    this.mask = null
  },
  methods: {
    hide() {
      this.$emit('update:visible', false)
    },
    onEnter() {
      this.$emit('show')
      this.focus()
      if (this.autoZIndex) {
        ZIndex.set('modal', this.mask, this.baseZIndex || this.$primevue.config.zIndex.modal)
      }
    },
    onBeforeLeave() {
      if (this.modal) {
        addClass(this.mask, 'p-overlay-mask-leave')
      }
    },
    onLeave() {
      this.$emit('hide')
    },
    onAfterLeave() {
      if (this.autoZIndex) {
        ZIndex.clear(this.mask)
      }
      this.containerVisible = false
      this.$emit('after-hide')
    },
    focus() {
      const findFocusableElement = (container) => {
        return container && container.querySelector('[autofocus]')
      }

      let focusTarget = this.$slots.header && findFocusableElement(this.headerContainer)

      if (!focusTarget) {
        focusTarget = this.$slots.default && findFocusableElement(this.container)

        if (!focusTarget) {
          focusTarget = this.closeButton
        }
      }

      focusTarget && focus(focusTarget)
    },
    containerRef(el) {
      this.container = el
      if (!el) {
        return
      }
      if (this.position === 'full') {
        return
      }
      if (this.mask) {
        this.mask.style.height = `${this.container.offsetHeight}px`
        if (this.position === 'bottom') {
          this.mask.style.top = null
          this.mask.style.bottom = `0`
        }
      }
    },
    maskRef(el) {
      this.mask = el
      if (!el) {
        return
      }
      if (this.position === 'full') {
        return
      }
      if (this.container) {
        this.mask.style.height = `${this.container.offsetHeight}px`
        if (this.position === 'bottom') {
          this.mask.style.top = null
          this.mask.style.bottom = `0`
        }
      }
    },
    contentRef(el) {
      this.content = el
    },
    headerContainerRef(el) {
      this.headerContainer = el
    },
    closeButtonRef(el) {
      this.closeButton = el ? el.$el : undefined
    },
  },
  computed: {
    fullScreen() {
      return this.position === 'full'
    },
    closeAriaLabel() {
      return this.$primevue.config.locale.aria ? this.$primevue.config.locale.aria.close : undefined
    }
  },
  directives: {
    focustrap: FocusTrap
  },
  components: {
    Button,
    Portal,
    TimesIcon
  }
}
</script>
