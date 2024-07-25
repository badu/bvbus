<template>
  <div :class="classes" :style="getStyles()">
    <div class="card" @click="handleOnClick">
      <div class="icon">
        <i :class="icon"></i>
      </div>
      <div v-if="selected" class="content">
        <div class="title">
          <h1>{{ title }}</h1>
        </div>
        <div class="text">
          <p>{{ text }}</p>
        </div>
        <button type="button" class="close-button" @click="handleClose">
          <i class="pi pi-times"></i>
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import {defineComponent, computed} from 'vue'

export default defineComponent({
  name: 'CardComponent',
  props: {
    id: Number,
    index: Number,
    icon: String,
    title: String,
    text: String,
    selected: Boolean,
    select: Function
  },
  setup(props) {
    const getStyles = () => {
      let styles = {
        left: `calc(${props.index * 20}% - ${props.index * 20}px)`,
        zIndex: props.index
      }

      if (props.selected) {
        styles.left = '50%'
        styles.zIndex = 10
      }

      return styles
    }

    const handleOnClick = (event) => {
      event.stopImmediatePropagation()
      if (!props.selected) {
        props.select(props.id)
      }
    }

    const handleClose = (event) => {
      event.stopImmediatePropagation()
      if (props.selected) {
        props.select(null)
      }
    }

    const classes = computed(() => ({'card-wrapper': true, selected: props.selected}))

    return {
      getStyles,
      handleOnClick,
      handleClose,
      classes
    }
  }
})
</script>
<style lang="scss" scoped>
@import "../assets/cards.scss";
</style>