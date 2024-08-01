<script setup>
import {ref, computed, onMounted, useCssModule} from "vue"

const props = defineProps({
  height: {
    type: Number,
    default: () => {
      return 40
    }
  },
  id: {
    type: String,
    required: true,
    default: () => {
      return "marquee-id"
    },
  },
  paused: {
    type: Boolean,
    default: () => {
      return false
    },
  },
  repeat: {
    type: Number,
    default: () => {
      return 100
    },
  },
  reverse: {
    type: Boolean,
    default: () => {
      return false
    },
  },
  space: {
    type: Number,
    default: () => {
      return 10
    },
  },
  speed: {
    type: Number,
    default: () => {
      return 10000
    },
  },
  vertical: {
    type: Boolean,
    default: () => {
      return false
    }
  },
  width: {
    type: Number,
    default: () => {
      return 100
    },
  },
})

let container = ref(null)
let containerWidth = ref(0)
let items = ref([])
let itemsLength = ref(0)
const itemsWidth = ref([])
let style = ref(null)

const styleElement = computed(
    () =>
        `
        animation-duration: ${props.speed}ms;
        animation-direction: ${props.reverse ? "reverse" : "normal"};
        animation-play-state: ${props.paused ? "paused" : "running"};
        height: ${props.vertical ? props.height : '40px'}`
)

const calculateContainerWidth = () => {
  for (let index = 0; index < itemsLength; index++) {
    itemsWidth.value.push(items[index].offsetWidth)
    setItemSpace(index)
  }
  containerWidth = itemsWidth.value.reduce((a, b) => a + b, 0)
}

const cloneItems = () => {
  const repeatCounter = getRepeatCounter()
  for (let index = 0; index < repeatCounter; index++) {
    container.appendChild(items[index].cloneNode(true))
  }
}

const getRepeatCounter = () => {
  return items.length * props.repeat
}

const setItems = () => {
  items = container.children// get all children that will be put inside the slot of the component
}

const setItemsLength = () => {
  itemsLength = items.length
}

const setItemSpace = (index) => {
  props.vertical ? items[index].style.marginBottom = `${props.space}px` : items[index].style.marginRight = `${props.space}px`
}

const setContainer = () => {
  container = document.querySelector(`#${props.id} .${props.vertical ? style.sliderVerticalContainer : style.sliderContainer}`
  )
}

const setContainerWidth = () => {
  if (props.vertical) {
    container.style.width = 'auto'
  } else {
    container.style.width = '40vw'
  }
}

onMounted(() => {
  style = useCssModule()
  setContainer()
  setItems()
  setItemsLength()
  calculateContainerWidth()
  setContainerWidth()
  cloneItems()
})

</script>

<template>
  <div :id="id" :class="$style.slider">
    <div :class="vertical ? $style.sliderVerticalContainer : $style.sliderContainer" :style="styleElement">
      <slot>

      </slot>
    </div>
  </div>
</template>

<style lang="css" module>
.slider {
  overflow: hidden;
}

.sliderContainer {
  width: 100%;
  animation-name: animateHorizontal;
  animation-timing-function: linear;
  animation-iteration-count: infinite;
  display: flex;
}

.sliderVerticalContainer {
  height: fit-content;
  animation-name: animateVertical;
  animation-timing-function: linear;
  animation-iteration-count: infinite;
  display: flex;
  flex-direction: column;
}

@keyframes animateHorizontal {
  0% {
    transform: translateX(0%);
  }
  100% {
    transform: translateX(-100%);
  }
}

@keyframes animateVertical {
  0% {
    transform: translateY(0%);
  }
  100% {
    transform: translateY(-100%);
  }
}
</style>
