<template>
  <div class="marquee-container">
    <div
        v-for="(item, index) in items"
        :key="index"
        v-show="currentItem === index"
        class="marquee-item">
      <Tag
          :rounded="true"
          :value="item.n"
          :style="{ minWidth: '40px',maxWidth:'40px', userSelect: 'none', fontFamily: 'TheLedDisplaySt', backgroundColor: item.c, color:item.tc }"/>
        <span class="text" ref="text">{{ item.f }} - {{ item.t }}</span>
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
      currentItem: 0,
      marqueeWidth: 0,
      running: false,
    }
  },
  mounted() {
    this.running = true
    this.startMarquee()
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
        if (this.$refs.text && this.$refs.text.length === 1) {
          this.marqueeWidth = this.$refs.text[0].offsetWidth
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
      this.currentItem = (this.currentItem + 1) % this.items.length
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
  animation: marquee 4s linear infinite;
  font-weight: 800;
}

@keyframes marquee {
  from {
    transform: translateX(10%);
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
</style>
