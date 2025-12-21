<template>
  <div class="sticky-banner">
    <div class="banner-content">
      <p>
        Launched your App and need server / container monitoring? To support my work checkout: 
        <a href="https://bareagent.io" target="_blank" rel="noopener noreferrer">https://bareagent.io</a>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.sticky-banner {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  width: 100%;
  background: linear-gradient(135deg, #0077b6 0%, #023e8a 100%);
  color: white;
  padding: 1rem 1rem;
  text-align: center;
  z-index: 1000;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

.dark .sticky-banner {
  background: linear-gradient(135deg, #023e8a 0%, #03045e 100%);
}

.banner-content {
  max-width: 1200px;
  margin: 0 auto;
}

.banner-content p {
  margin: 0;
  font-size: 0.875rem;
  line-height: 1.4;
}

.banner-content a {
  color: #90e0ef;
  text-decoration: underline;
  font-weight: 600;
  transition: color 0.2s ease;
}

.banner-content a:hover {
  color: #caf0f8;
}

.dark .banner-content a {
  color: #48cae4;
}

.dark .banner-content a:hover {
  color: #90e0ef;
}

@media (max-width: 768px) {
  .sticky-banner {
    padding: 0.4rem 0.75rem;
  }
  
  .banner-content p {
    font-size: 0.8rem;
  }
}
</style>
