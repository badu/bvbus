<template>
  <div class="marquee-container">
    <div
        v-for="(item, index) in items"
        :key="index"
        v-show="currentIndex === index"
        class="marquee-item">
      <Tag
          :rounded="true"
          :value="item.n"
          :style="{ minWidth: '40px',maxWidth:'40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: item.c, color:item.tc }"/>
      <div ref="mask" class="marquee-mask" :style="{ width: marqueeWidth > 0 ? marqueeWidth + 'px' : '100%' }">
        <div ref="text" class="text" >{{ item.f }} - {{ item.t }}</div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    items: {
      type: Array,
      required: true,
    },
  },
  data() {
    return {
      currentIndex: 0,
      marqueeWidth: 0,
      running: false,
      hasInitialWidth: false,
      marqueeInitialWidth: 0,
    }
  },
  mounted() {
    this.$nextTick(() => {
      this.running = true
      this.startMarquee()
    })
  },
  unmounted() {
    this.running = false
  },

  methods: {
    startMarquee() {
      this.setMarqueeWidth()
      this.animateMarquee()
    },
    setMarqueeWidth() {
      if (this.running) {
        if (this.$refs) {
          if (!this.hasInitialWidth) {
            this.marqueeInitialWidth = this.$refs.mask[0].offsetWidth
            this.hasInitialWidth = true
          }
          this.marqueeWidth = this.$refs.text[this.currentIndex % this.items.length].offsetWidth
        }
      }
    },
    animateMarquee() {
      if (!this.running) {
        return
      }

      setTimeout(() => {
        this.nextItem()
      }, 5000)
    },
    nextItem() {
      if (!this.running) {
        return
      }
      this.currentIndex = (this.currentIndex + 1) % this.items.length
      setTimeout(this.startMarquee, 500)
    },
  },
}
</script>

<style scoped>
.marquee-container {
  position: relative;
  overflow: hidden;
  height: 42px;
}

.marquee-item {
  position: absolute;
  top: 0;
  left: 0;
  display: flex;
  align-items: center;
  white-space: nowrap;
  width: 100%;
  animation: slideIn 0.5s ease-in-out;
  text-align: center;
  vertical-align: center;
}

.text {
  display: inline-block;
  white-space: nowrap;
  color: #FED053;
  font-weight: 800;
  animation: marquee 4s linear infinite;
}

@keyframes marquee {
  from {
    transform: translateX(50%);
  }
  to {
    transform: translateX(-100%);
  }
}

@keyframes slideIn {
  from {
    transform: translateY(-100%);
  }
  to {
    transform: translateY(0);
  }
}

.marquee-mask {
  overflow: hidden;
  white-space: nowrap;
}
</style>
