<script>
import BaseComponent from '@primevue/core/basecomponent'
import DrawerStyle from '@/components/drawer.style.ts'

export default {
  name: 'BaseDrawer',
  extends: BaseComponent,
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    position: {
      type: String,
      default: 'full'
    },
    header: {
      type: null,
      default: null
    },
    baseZIndex: {
      type: Number,
      default: 0
    },
    autoZIndex: {
      type: Boolean,
      default: true
    },
    showCloseIcon: {
      type: Boolean,
      default: true
    },
    closeButtonProps: {
      type: Object,
      default: () => {
        return {severity: 'secondary', text: true, rounded: true};
      }
    },
    closeIcon: {
      type: String,
      default: undefined
    },
    modal: {
      type: Boolean,
      default: true
    },
    hasContent: {
      type: Boolean,
      default: true
    }
  },
  style: DrawerStyle,
  provide() {
    return {
      $pcDrawer: this,
      $parentInstance: this
    };
  }
};
</script>
